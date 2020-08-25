package controllers

import (
	"fmt"
	"github.com/muerwre/vault-golang/constants"
	"github.com/muerwre/vault-golang/request"
	"github.com/muerwre/vault-golang/response"
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

		nc.DB.NodeViewRepository.UpdateView(uid, node.ID)
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
	uid := c.MustGet("UID").(uint)

	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": codes.IncorrectData})
		return
	}

	if params.Take == 0 {
		params.Take = 40
	}

	before := &[]models.Node{}
	after := &[]models.Node{}
	heroes := &[]models.Node{}
	updated := &[]models.Node{}
	recent := &[]models.Node{}

	valid := []uint{}

	q := nc.DB.Preload("User").Preload("User.Photo").Model(&models.Node{})
	// TODO: move to repo
	nc.DB.NodeRepository.WhereIsFlowNode(q)

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		q.Where("created_at > ?", params.Start).
			Order("created_at DESC").
			Find(&before)

		q.Where("created_at < ?", params.End).
			Order("created_at DESC").
			Offset(0).
			Limit(params.Take).
			Find(&after)

		if params.WithHeroes {
			q.Order("RAND()").
				Where("type = ? AND is_heroic = ?", "image", true).
				Offset(0).
				Limit(20).
				Find(&heroes)
		}

		wg.Done()
	}()

	go func() {
		if uid != 0 && params.WithUpdated {
			q.Order("created_at DESC").
				Joins("LEFT JOIN node_view AS node_view ON node_view.nodeId = node.id AND node_view.userId = ?", uid).
				Where("node_view.visited < node.commented_at").
				Limit(10).
				Find(&updated)
		}

		exclude := make([]uint, len(*updated)+1)
		exclude[0] = 0

		for k, v := range *updated {
			exclude[k+1] = v.ID
		}

		if params.WithRecent {
			q.Order("created_at DESC").
				Where("commented_at IS NOT NULL AND id NOT IN (?)", exclude).
				Limit(16).
				Find(&recent)
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

	resp := new(response.FlowResponse).Init(*before, *after, *heroes, *updated, *recent, valid)

	c.JSON(http.StatusOK, resp)
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

	if err != nil || node.Type == "" || !node.CanBeCommented() {
		if err != nil {
			logrus.Warnf(err.Error())
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": codes.NodeNotFound})
		return
	}

	if err = c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	comment, err := nc.LoadCommentFromData(data.ID, node, u)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lostFiles, err := nc.UpdateCommentFiles(data, comment)

	if err != nil {
		logrus.Warnf("Can't load node files while updating comment %d for node %d: %s", node.ID, comment.ID, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.CantSaveComment})
		return
	}

	if err := nc.UpdateCommentText(data, comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err = nc.DB.NodeRepository.SaveCommentWithFiles(comment); err != nil {
		logrus.Warnf("Failed to save comment %d for node: %s", comment.ID, node.ID, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.CantSaveComment})
		return
	}

	nc.UnsetFilesTarget(lostFiles)
	nc.UpdateBriefFromComment(node, comment)
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
	user := c.MustGet("User").(*models.User)

	if !user.CanCreateNode() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": codes.NotEnoughRights})
		return
	}

	data := models.Node{}

	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	node, err := nc.LoadNodeFromData(data, user)

	if err != nil {
		logrus.Warnf("Can't load node from data: %s\nData:\n%+v", err.Error(), data)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Update previous node cover target to be null if its changed
	if err = nc.UpdateNodeCoverIfChanged(data, node); err != nil {
		logrus.Warnf("Can't load node cover from data: %s\nData:\n%+v", err.Error(), data)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	lostFiles, err := nc.UpdateNodeFiles(data, node)

	if err != nil {
		logrus.Warnf("Can't load node files while updating node %d: %s", data.ID, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nc.UpdateNodeTitle(data, node)

	if err = nc.UpdateNodeBlocks(data, node); err != nil {
		logrus.Warnf("Received suspicious blocks while updating node %d: %s", data.ID, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	node.UpdateDescription()
	node.UpdateThumbnail()

	if err = nc.DB.NodeRepository.SaveNodeWithFiles(node); err != nil {
		logrus.Warnf("Failed to save node %d: %s", node.ID, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	nc.UpdateFilesMetadata(data.Files, node.Files)
	nc.SetFilesTarget(node.FilesOrder, "node")
	nc.UnsetFilesTarget(lostFiles)
	nc.UnsetNodeCoverTarget(data, node)

	c.JSON(http.StatusOK, gin.H{"node": node})
}

func (nc NodeController) UpdateBriefFromComment(node *models.Node, comment *models.Comment) {
	if node.Description == "" && *comment.UserID == *node.UserID && len(comment.Text) >= 64 {
		node.Description = comment.Text
		nc.DB.Save(&node)
	}
}

// TODO: Move everything below this line to usecase:

func (nc NodeController) UpdateCommentFiles(data *models.Comment, comment *models.Comment) ([]uint, error) {
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

		comment.Files = make([]*models.File, 0)

		query := nc.DB.
			Order(gorm.Expr(fmt.Sprintf("FIELD(id, %s)", ids))).
			Find(
				&comment.Files,
				"id IN (?) AND TYPE IN (?)",
				[]uint(data.FilesOrder),
				structs.Names(models.COMMENT_FILE_TYPES),
			)

		if query.Error != nil {
			return nil, query.Error
		}

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

	return lostFiles, nil
}

func (nc *NodeController) SetFilesTarget(files []uint, target string) {
	if len(files) > 0 {
		nc.DB.Model(&models.File{}).Where("id IN (?)", []uint(files)).Update("target", target)
	}
}

func (nc *NodeController) UnsetFilesTarget(files []uint) {
	if len(files) > 0 {
		nc.DB.Model(&models.File{}).Where("id IN (?)", files).Update("target", nil)
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
		nc.DB.Preload("User").Preload("User.Photo").First(&comment, "id = ?", id)
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
		if df == nil || df.Type != constants.FileTypeAudio {
			continue
		}

		for _, cf := range comment {
			if cf != nil && cf.ID == df.ID && cf.Metadata.Title != df.Metadata.Title {
				cf.Metadata.Title = df.Metadata.Title

				if err := nc.DB.FileRepository.UpdateMetadata(cf, cf.Metadata); err != nil {
					logrus.Warnf("Can't update file metadata %d: %s", cf.ID, err.Error())
				}

				break
			}
		}
	}
}

func (nc NodeController) UpdateNodeCoverIfChanged(data models.Node, node *models.Node) error {
	// Validate node cover
	if data.Cover != nil && data.Cover.ID != 0 {
		query := nc.DB.Model(&models.File{}).Where("id = ?", data.Cover.ID).First(&node.Cover)

		if query.Error != nil {
			return query.Error
		}

		*node.CoverID = data.Cover.ID
	}

	return nil
}

func (nc NodeController) UpdateNodeTitle(data models.Node, node *models.Node) {
	node.Title = data.Title

	if len(node.Title) > 64 {
		node.Title = node.Title[:64]
	}
}

func (nc NodeController) UpdateNodeBlocks(data models.Node, node *models.Node) error {
	node.ApplyBlocks(data.Blocks)

	if val, ok := validation.NodeValidators[node.Type]; ok {
		err := val(node)

		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf(codes.UnknownNodeType)
	}

	return nil
}

func (nc NodeController) LoadNodeFromData(data models.Node, u *models.User) (*models.Node, error) {
	node := &models.Node{}

	if data.ID != 0 {
		nc.DB.Preload("User").Preload("User.Photo").First(&node, "id = ?", data.ID)
		node.Cover = nil
	} else {
		node.User = u
		node.UserID = &u.ID
		node.Type = data.Type
		node.IsPublic = true
		node.IsPromoted = true
		node.Tags = make([]*models.Tag, 0)
	}

	if node.Type == "" || !models.FLOW_NODE_TYPES.Contains(node.Type) {
		return nil, fmt.Errorf(codes.IncorrectType)
	}

	if !node.CanBeEditedBy(u) {
		return nil, fmt.Errorf(codes.NotEnoughRights)
	}

	return node, nil
}

func (nc NodeController) UpdateNodeFiles(data models.Node, node *models.Node) ([]uint, error) {
	// Finding out valid comment attaches and sorting them according to files_order
	originFiles := make([]uint, len(node.FilesOrder))
	copy(originFiles, node.FilesOrder)

	// Setting FilesOrder based on sorted Files array of input data
	data.FilesOrder = make(models.CommaUintArray, 0)

	for _, v := range data.Files {
		if v == nil {
			continue
		}

		data.FilesOrder = append(data.FilesOrder, v.ID)
	}

	if len(data.FilesOrder) > 0 {
		ids, _ := data.FilesOrder.Value()

		data.Files = make([]*models.File, 0)

		query := nc.DB.
			Order(gorm.Expr(fmt.Sprintf("FIELD(id, %s)", ids))).
			Find(&data.Files, "id IN (?)", []uint(data.FilesOrder))

		if query.Error != nil {
			return nil, query.Error
		}

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

	return lostFiles, nil
}

func (nc NodeController) UnsetNodeCoverTarget(data models.Node, node *models.Node) {
	if node.Cover != nil && (data.Cover == nil || data.Cover.ID == 0 || data.Cover.ID != *node.CoverID) {
		nc.UnsetFilesTarget([]uint{*node.CoverID})
	}
}
