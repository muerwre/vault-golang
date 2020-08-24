package controllers

import (
	"fmt"
	"github.com/muerwre/vault-golang/constants"
	"github.com/muerwre/vault-golang/request"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
	"github.com/muerwre/vault-golang/utils/validation"
)

type NodeController struct {
	DB db.DB
}

// GetNode /node:id - returns single node with tags, likes count and files
func (nc *NodeController) GetNode(c *gin.Context) {
	uid := c.MustGet("UID").(uint)
	u := c.MustGet("User").(*models.User)
	d := nc.DB

	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NodeNotFound})
		return
	}

	node, err := nc.DB.NodeRepository.GetFullNode(id, u.Role == models.USER_ROLES.ADMIN, u.ID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NodeNotFound})
		return
	}

	if uid != 0 {
		node.IsLiked = d.NodeRepository.IsNodeLikedBy(node, uid)
	}

	node.LikeCount = d.NodeRepository.GetNodeLikeCount(node)
	node.Files, _ = nc.DB.FileRepository.GetFilesByIds([]uint(node.FilesOrder))

	node.SortFiles()

	c.JSON(http.StatusOK, gin.H{"node": node})
}

// GetNodeComments /node/:id/comments - returns comments for node
func (nc *NodeController) GetNodeComments(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NodeNotFound})
		return
	}

	take, err := strconv.Atoi(c.Query("take"))

	if err != nil {
		take = 100
	}

	skip, err := strconv.Atoi(c.Query("skip"))

	if err != nil {
		skip = 0
	}

	order := c.Query("order")

	if order != "ASC" {
		order = "DESC"
	}

	comments, count := nc.DB.NodeRepository.GetComments(id, take, skip, order)

	c.JSON(http.StatusAccepted, gin.H{"comments": comments, "comment_count": count})
}

// GetDiff /nodes/diff gets newer and older nodes
func (nc *NodeController) GetDiff(c *gin.Context) {
	params := &request.NodeDiffParams{}
	err := c.Bind(&params)
	d := nc.DB
	uid := c.MustGet("UID").(uint)

	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": codes.IncorrectData})
		return
	}

	if params.Take == 0 {
		params.Take = 40
	}

	before := &[]models.FlowNode{}
	after := &[]models.FlowNode{}
	heroes := &[]models.FlowNode{}
	updated := &[]models.FlowNode{}
	recent := &[]models.FlowNode{}
	valid := []uint{}

	// TODO: move to repo
	q := nc.DB.NodeRepository.WhereIsFlowNode(d.Model(&models.Node{}))

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		q.Where("created_at > ?", params.Start).
			Order("created_at DESC").
			Scan(&before)

		q.Where("created_at < ?", params.End).
			Order("created_at DESC").
			Offset(0).
			Limit(params.Take).Scan(&after)

		if params.WithHeroes {
			d.Model(&models.Node{}).
				Where("type = ? AND is_heroic = ?", "image", true).
				Order("RAND()").
				Offset(0).
				Limit(20).
				Scan(&heroes)
		}

		wg.Done()
	}()

	go func() {
		if uid != 0 && params.WithUpdated {
			q.Order("created_at DESC").
				Joins("LEFT JOIN node_view AS node_view ON node_view.nodeId = node.id AND node_view.userId = ?", uid).
				Where("node_view.visited < node.commented_at").
				Limit(10).
				Scan(&updated)
		}

		exclude := make([]uint, len(*updated)+1)
		exclude[0] = 0

		for k, v := range *updated {
			exclude[k+1] = v.ID
		}

		if params.WithRecent {
			q.Order("created_at DESC").
				Preload("User").
				Where("commented_at IS NOT NULL AND id NOT IN (?)", exclude).
				Limit(16).
				Scan(&recent)
		}

		if params.WithValid {
			rows, err := q.Table("node").
				Select("id").
				Where("created_at >= ? AND created_at <= ?", params.End, params.Start).
				Rows()

			if err == nil {
				id := uint(0)
				for i := 0; rows.Next(); i += 1 {
					err = rows.Scan(&id)

					if id > 0 && err == nil {
						valid = append(valid, id)
					}
				}
			}
		}

		wg.Done()
	}()

	wg.Wait()

	c.JSON(http.StatusOK, gin.H{
		"before":  before,
		"after":   after,
		"heroes":  heroes,
		"updated": updated,
		"recent":  recent,
		"valid":   valid,
	})
}

func (nc *NodeController) LockComment(c *gin.Context) {
	d := nc.DB
	u := c.MustGet("User").(*models.User)
	cid := c.Param("cid")
	params := request.NodeLockCommentRequest{}

	err := c.BindJSON(&params)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	comment := &models.Comment{}
	d.Unscoped().Where("id = ?", cid).First(&comment)

	if comment == nil || comment.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.CommentNotFound})
		return
	}

	if !u.CanEditComment(comment) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": codes.NotEnoughRights})
		return
	}

	if params.IsLocked {
		d.Delete(&comment)
	} else {
		comment.DeletedAt = nil
		d.Model(&comment).Unscoped().Where("id = ?", comment.ID).Update("deletedAt", nil)
	}

	c.JSON(http.StatusOK, gin.H{"deteled_at": &comment.DeletedAt})
}

// PostComment - /node/:id/comment saving and creating comments
func (nc *NodeController) PostComment(c *gin.Context) {
	data := &models.Comment{}
	u := c.MustGet("User").(*models.User)
	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if nid == 0 || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.NodeNotFound})
		return
	}

	node, err := nc.DB.NodeRepository.GetById(uint(nid))

	if err != nil {
		logrus.Warnf(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.NodeNotFound})
		return
	}

	if node.Type == "" || !node.CanBeCommented() {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.NodeNotFound})
		return
	}

	err = c.BindJSON(&data)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	comment, err := nc.LoadCommentFromData(data.ID, node, u)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: unset comment and node files in defer
	nc.UpdateCommentFiles(data, comment)

	if err := nc.UpdateCommentText(data, comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Updating node brief
	nc.UpdateBriefFromComment(node, comment)

	if comment, err = nc.DB.NodeRepository.SaveCommentWithFiles(comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.CantSaveComment})
		return
	}

	nc.UpdateFilesMetadata(data.Files, comment.Files)

	c.JSON(http.StatusOK, gin.H{"comment": comment})
}

// PostTags - POST /node/:id/tags - updates node tags
func (nc *NodeController) PostTags(c *gin.Context) {
	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	d := nc.DB
	u := c.MustGet("User").(*models.User)

	node := &models.Node{}
	tags := []*models.Tag{}

	params := request.NodeTagsPostRequest{}

	if nid == 0 || err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NodeNotFound})
		return
	}

	d.First(&node, "id = ?", nid)

	if node == nil || node.ID == 0 || !node.CanBeTaggedBy(u) {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NotEnoughRights})
		return
	}

	err = c.BindJSON(&params)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	for i := 0; i < len(params.Tags); i += 1 {
		params.Tags[i] = strings.ToLower(params.Tags[i])
	}

	d.Where("title IN (?)", params.Tags).Find(&tags)

	for _, v := range params.Tags {
		if !models.TagArrayContains(tags, v) {
			tag := models.Tag{Title: v}
			d.Set("gorm:association_autoupdate", false).Save(&tag)
			tags = append(tags, &tag)
		}
	}

	d.Model(&node).Association("Tags").Replace(tags)

	c.JSON(http.StatusOK, gin.H{"node": node})
}

// PostLike - POST /node/:id/like - likes or dislikes node
func (nc NodeController) PostLike(c *gin.Context) {
	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	d := nc.DB
	u := c.MustGet("User").(*models.User)

	node := &models.Node{}

	if nid == 0 || err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NodeNotFound})
		return
	}

	d.First(&node, "id = ?", nid)

	if node == nil || node.ID == 0 || !node.CanBeLiked() {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NotEnoughRights})
		return
	}

	isLiked := d.NodeRepository.IsNodeLikedBy(node, u.ID)

	if isLiked {
		d.Model(&node).Association("Likes").Delete(u)
	} else {
		d.Model(&node).Association("Likes").Append(u)
	}

	c.JSON(http.StatusOK, gin.H{"is_liked": !isLiked})
}

// PostLock - POST /node/:id/lock - safely deletes node
func (nc NodeController) PostLock(c *gin.Context) {
	d := nc.DB
	u := c.MustGet("User").(*models.User)
	params := request.NodeLockRequest{}

	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	err = c.BindJSON(&params)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	node := &models.Node{}

	d.Unscoped().First(&node, "id = ?", nid)

	if node == nil || node.ID == 0 || !node.CanBeEditedBy(u) {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NotEnoughRights})
		return
	}

	if params.IsLocked {
		d.Unscoped().Model(&node).Update("deleted_at", time.Now().Truncate(time.Second))
	} else {
		d.Unscoped().Model(&node).Update("deleted_at", nil)
	}

	c.JSON(http.StatusOK, gin.H{"deleted_at": node.DeletedAt})
}

// PostHeroic - POST /node/:id/heroic - sets heroic status to node
func (nc NodeController) PostHeroic(c *gin.Context) {
	d := nc.DB
	u := c.MustGet("User").(*models.User)

	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if err != nil || nid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	node := &models.Node{}

	d.First(&node, "id = ?", nid)

	if node == nil || node.ID == 0 || !node.CanBeHeroedBy(u) {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NotEnoughRights})
		return
	}

	d.Model(&node).Update("is_heroic", !node.IsHeroic)

	c.JSON(http.StatusOK, gin.H{"is_heroic": node.IsHeroic})
}

// PostCellView - POST /node/:id/cell-view - sets cel display for node
func (nc NodeController) PostCellView(c *gin.Context) {
	d := nc.DB
	u := c.MustGet("User").(*models.User)
	params := request.NodeCellViewPostRequest{}

	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	err = c.BindJSON(&params)

	if err != nil || !models.NODE_FLOW_DISPLAY.Contains(params.Flow.Display) {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	node := &models.Node{}

	if nid == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NodeNotFound})
		return
	}

	d.First(&node, "id = ?", nid)

	if node == nil || node.ID == 0 || !node.CanBeEditedBy(u) {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NotEnoughRights})
		return
	}

	node.Flow = params.Flow

	d.Model(&node).Update("flow", node.Flow)

	c.JSON(http.StatusOK, gin.H{"flow": node.Flow})
}

// GetRelated - GET /node/:id/related - gets related nodes
func (nc NodeController) GetRelated(c *gin.Context) {
	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if err != nil || nid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	related, err := nc.DB.NodeRepository.GetRelated(uint(nid))
	c.JSON(http.StatusOK, gin.H{"related": related})
}

func (nc NodeController) PostNode(c *gin.Context) {
	d := nc.DB
	u := c.MustGet("User").(*models.User)

	if !u.CanCreateNode() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": codes.NotEnoughRights})
		return
	}

	params := request.NodePostRequest{}

	err := c.BindJSON(&params)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	data := params.Node
	node := &models.Node{}

	if data.ID != 0 {
		d.First(&node, "id = ?", data.ID)
		node.Cover = nil
	} else {
		node.User = u
		node.UserID = &u.ID
		node.Type = data.Type
		node.IsPublic = true
		node.IsPromoted = true
		node.Tags = make([]*models.Tag, 0)
	}

	if data.Type == "" || !models.FLOW_NODE_TYPES.Contains(data.Type) {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectType})
		return
	}

	if !node.CanBeEditedBy(u) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": codes.NotEnoughRights})
		return
	}

	// Update previous node cover target to be null if its changed
	nc.UpdateNodeCoverIfChanged(&data, node)

	// Clear previous cover target if succeeded
	defer func() {
		if node.Cover != nil && (data.Cover == nil || data.Cover.ID == 0 || data.Cover.ID != *node.CoverID) {
			nc.DB.Model(&node.Files).Where("id IN (?)", node.CoverID).Update("target", nil)
		}
	}()

	// Finding out valid comment attaches and sorting them according to files_order
	originFiles := make([]uint, len(node.FilesOrder))
	copy(originFiles, node.FilesOrder)

	// Setting FilesOrder based on sorted Files array of input data
	data.FilesOrder = make(models.CommaUintArray, 0)

	for _, v := range data.Files {
		if v == nil {
			continue
		}

		data.FilesOrder = append(node.FilesOrder, v.ID)
	}

	if len(data.FilesOrder) > 0 {
		ids, _ := data.FilesOrder.Value()

		d.Order(gorm.Expr(fmt.Sprintf("FIELD(id, %s)", ids))).
			Find(&data.Files, "id IN (?)", []uint(data.FilesOrder))

		node.ApplyFiles(data.Files)
	} else {
		node.Files = make([]*models.File, 0)
		node.FilesOrder = make(models.CommaUintArray, 0)
	}

	// Detecting lost files
	lostFiles := make(models.CommaUintArray, 0)

	for _, v := range originFiles {
		if !node.FilesOrder.Contains(v) {
			lostFiles = append(lostFiles, v)
		}
	}

	// Unsetting them
	if len(lostFiles) > 0 {
		d.Model(&node.Files).Where("id IN (?)", []uint(lostFiles)).Update("target", nil)
	}

	nc.UpdateNodeTitle(&data, node)

	node.ApplyBlocks(data.Blocks)

	if val, ok := validation.NodeValidators[node.Type]; ok {
		err = val(node)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// Update node metadata
	node.UpdateDescription()
	node.UpdateThumbnail()

	nc.UpdateFilesMetadata(data.Files, node.Files)

	// Save node and its files
	d.Set("gorm:association_autoupdate", false).
		Set("gorm:association_save_reference", false).
		Save(&node).
		Association("Files").Replace(node.Files)

	// Node not saved
	if node.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	// Updating current files target
	if len(node.FilesOrder) > 0 {
		d.Model(&models.File{}).Where("id IN (?)", node.FilesOrder).Update("Target", "node")
	}

	c.JSON(http.StatusOK, gin.H{"node": node})
}

func (nc NodeController) UpdateBriefFromComment(node *models.Node, comment *models.Comment) {
	if node.Description == "" && *comment.UserID == *node.UserID && len(comment.Text) >= 64 {
		node.Description = comment.Text
		nc.DB.Save(&node)
	}
}

// TODO: Move everything below this line to usecase:

func (nc NodeController) UpdateCommentFiles(data *models.Comment, comment *models.Comment) {
	// Setting FilesOrder based on sorted Files array of input data
	data.FilesOrder = make(models.CommaUintArray, 0)

	for _, v := range data.Files {
		data.FilesOrder = append(data.FilesOrder, v.ID)
	}

	// Finding out valid comment attaches and sorting them according to files_order
	originFiles := make([]uint, len(comment.FilesOrder))
	copy(originFiles, comment.FilesOrder)

	lostFiles := make(models.CommaUintArray, 0)
	comment.FilesOrder = make(models.CommaUintArray, 0)

	// Loading that files
	if len(data.FilesOrder) > 0 {
		ids, _ := data.FilesOrder.Value()

		nc.DB.Order(gorm.Expr(fmt.Sprintf("FIELD(id, %s)", ids))).
			Find(&comment.Files, "id IN (?) AND TYPE IN (?)", []uint(data.FilesOrder), structs.Names(models.COMMENT_FILE_TYPES))

		for i := 0; i < len(comment.Files); i += 1 { // TODO: limit files count
			comment.FilesOrder = append(comment.FilesOrder, comment.Files[i].ID)
		}
	} else {
		comment.Files = make([]*models.File, 0)
		comment.FilesOrder = make(models.CommaUintArray, 0)
	}

	// Detecting lost files
	for _, v := range originFiles {
		if !comment.FilesOrder.Contains(v) {
			lostFiles = append(lostFiles, v)
		}
	}

	// Unsetting them
	if len(lostFiles) > 0 {
		nc.DB.Model(&comment.Files).Where("id IN (?)", []uint(lostFiles)).Update("target", nil)
	}
}

func (nc *NodeController) UpdateCommentText(data *models.Comment, comment *models.Comment) error {
	comment.Text = data.Text

	if len(comment.Text) > 2048 {
		comment.Text = comment.Text[0:2048]
	}

	if len(comment.Text) < 1 && len(comment.FilesOrder) == 0 {
		return fmt.Errorf(codes.TextRequired)
	}

	return nil
}

func (nc *NodeController) LoadCommentFromData(id uint, node *models.Node, user *models.User) (*models.Comment, error) {
	comment := &models.Comment{
		Files: make([]*models.File, 0),
	}

	if id != 0 {
		nc.DB.First(&comment, "id = ?", id)
	} else {
		comment.Node = node
		comment.NodeID = &node.ID
		comment.User = user
		comment.UserID = &user.ID
	}

	if *comment.NodeID != node.ID || !comment.CanBeEditedBy(user) {
		return nil, fmt.Errorf(codes.NotEnoughRights)
	}

	return comment, nil
}

func (nc NodeController) UpdateFilesMetadata(data []*models.File, comment []*models.File) {
	for _, df := range data {
		if df.Type != constants.FileTypeAudio {
			continue
		}

		for _, cf := range comment {
			if cf.ID == df.ID && cf.Metadata.Title != df.Metadata.Title {
				cf.Metadata.Title = df.Metadata.Title
				nc.DB.FileRepository.UpdateMetadata(cf, cf.Metadata)
				break
			}
		}
	}
}

func (nc NodeController) UpdateNodeCoverIfChanged(data *models.Node, node *models.Node) {
	// Validate node cover
	if data.Cover != nil && data.Cover.ID != 0 {
		nc.DB.First(&models.File{}, "id = ?", data.Cover.ID).Scan(&node.Cover)
		*node.CoverID = data.Cover.ID
	}
}

func (nc NodeController) UpdateNodeTitle(data *models.Node, node *models.Node) {
	node.Title = data.Title

	if len(node.Title) > 64 {
		node.Title = node.Title[:64]
	}
}
