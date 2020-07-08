package controllers

import (
	"fmt"
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

type NodeDiffParams struct {
	Start       time.Time `json:"start" form:"start"`
	End         time.Time `json:"end" form:"end"`
	Take        uint      `json:"take" form:"take"`
	WithHeroes  bool      `json:"with_heroes" form:"with_heroes"`
	WithUpdated bool      `json:"with_updated" form:"with_updated"`
	WithRecent  bool      `json:"with_recent" form:"with_recent"`
	WithValid   bool      `json:"with_valid" form:"with_valid"`
}

type NodeRelatedResponse struct {
	Albums  map[string][]models.NodeRelatedItem `json:"albums"`
	Similar []models.NodeRelatedItem            `json:"similar"`
}

var FlowNodeCriteria = "is_promoted = 1 AND is_public = 1 AND type IN (?)"

type NodeController struct{}

var Node = &NodeController{}

// GetNode /node:id - returns single node with tags, likes count and files
func (a *NodeController) GetNode(c *gin.Context) {
	id := c.Param("id")
	uid := c.MustGet("UID").(uint)
	d := c.MustGet("DB").(*db.DB)
	u := c.MustGet("User").(*models.User)

	node := &models.Node{}

	q := d.Unscoped().
		Preload("Tags").
		Preload("User").
		Preload("Cover")

	if u != nil && u.Role == models.USER_ROLES.ADMIN {
		q.First(&node, "id = ?", id)
	} else {
		q.First(&node, "id = ? AND (deleted_at IS NULL OR userID = ?)", id, u.ID)
	}

	if node.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NODE_NOT_FOUND})
		return
	}

	files_chan := make(chan []*models.File)

	go func() {
		files := make([]*models.File, len(node.FilesOrder))
		d.Where("id IN (?)", []uint(node.FilesOrder)).Find(&files)
		files_chan <- files
	}()

	if uid != 0 {
		node.IsLiked = d.IsNodeLikedBy(node, uid)
	}

	node.LikeCount = d.GetNodeLikeCount(node)
	node.Files = <-files_chan

	node.SortFiles()

	c.JSON(http.StatusOK, gin.H{"node": node})
}

// GetNodeComments /node/:id/comments - returns comments for node
func (a *NodeController) GetNodeComments(c *gin.Context) {
	d := c.MustGet("DB").(*db.DB)

	id := c.Param("id")

	take, err := strconv.Atoi(c.Query("take"))

	if err != nil {
		take = 100
	}

	skip, err := strconv.Atoi(c.Query("skip"))

	if err != nil {
		skip = 0
	}

	order := c.Query("order")

	if order != "DESC" {
		order = "ASC"
	}

	comments := &[]*models.Comment{}

	d.Preload("User").
		Preload("Files").
		Preload("User.Photo").
		Where("nodeId = ?", id).
		Offset(skip).
		Limit(take).
		Order(fmt.Sprintf("created_at %s", order)).
		Find(&comments)

	for _, v := range *comments {
		v.SortFiles()
	}

	c.JSON(http.StatusAccepted, gin.H{"comments": comments})
}

// GetDiff /nodes/diff gets newer and older nodes
func (_ *NodeController) GetDiff(c *gin.Context) {
	params := &NodeDiffParams{}
	err := c.Bind(&params)
	d := c.MustGet("DB").(*db.DB)
	uid := c.MustGet("UID").(uint)

	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": codes.INCORRECT_DATA})
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

	q := d.Model(&models.Node{}).Where(FlowNodeCriteria, structs.Values(models.FLOW_NODE_TYPES))

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
				Where("type = ? AND is_heroic < ?", "image", true).
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

func (_ *NodeController) LockComment(c *gin.Context) {
	d := c.MustGet("DB").(*db.DB)
	u := c.MustGet("User").(*models.User)
	cid := c.Param("cid")
	params := struct {
		IsLocked bool `json:"is_locked"`
	}{}

	err := c.BindJSON(&params)

	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.INCORRECT_DATA})
		return
	}

	comment := &models.Comment{}
	d.Unscoped().Where("id = ?", cid).First(&comment)

	if comment == nil || comment.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.COMMENT_NOT_FOUND})
		return
	}

	if !u.CanEditComment(comment) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": codes.NOT_ENOUGH_RIGHTS})
		return
	}

	if params.IsLocked {
		d.Delete(&comment)
	} else {
		comment.DeletedAt = nil
		d.Unscoped().Update(&comment)
	}

	c.JSON(http.StatusOK, gin.H{"deteled_at": &comment.DeletedAt})
}

// PostComment - /node/:id/comment savng and creating comments
func (_ *NodeController) PostComment(c *gin.Context) {
	comment := &models.Comment{}
	data := &models.Comment{}
	node := &models.Node{}
	d := c.MustGet("DB").(*db.DB)
	u := c.MustGet("User").(*models.User)
	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if nid == 0 || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.NODE_NOT_FOUND})
		return
	}

	d.First(&node, "id = ?", nid)

	if node.Type == "" || !node.CanBeCommented() {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.NODE_NOT_FOUND})
		return
	}

	err = c.BindJSON(&data)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.INCORRECT_DATA})
		return
	}

	if data.ID != 0 {
		d.First(&comment, "id = ?", data.ID)
	} else {
		comment.Node = node
		comment.NodeID = node.ID
		comment.User = u
		comment.UserID = u.ID
	}

	if comment.NodeID != node.ID || !comment.CanBeEditedBy(u) {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.NOT_ENOUGH_RIGHTS})
		return
	}

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

		d.Order(gorm.Expr(fmt.Sprintf("FIELD(id, %s)", ids))).
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
		d.Model(&comment.Files).Where("id IN (?)", []uint(lostFiles)).Update("target", nil)
	}

	comment.Text = data.Text

	if len(comment.Text) > 2048 {
		comment.Text = string(comment.Text[0:2048])
	}

	if len(comment.Text) < 2 && len(comment.FilesOrder) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.TEXT_REQUIRED})
		return
	}

	// Updating node brief
	if node.Description == "" && comment.UserID == node.UserID && len(comment.Text) >= 64 {
		node.Description = comment.Text
		d.Save(&node)
	}

	d.Set("gorm:association_autoupdate", false).
		Set("gorm:association_save_reference", false).
		Save(&comment).Association("Files").
		Replace(comment.Files)

	// Updating current files target
	if len(comment.FilesOrder) > 0 {
		d.Model(&comment.Files).Where("id IN (?)", []uint(comment.FilesOrder)).Update("Target", "comment")
	}

	c.JSON(http.StatusOK, gin.H{"comment": comment})
}

// PostTags - POST /node/:id/tags - updates node tags
func (_ *NodeController) PostTags(c *gin.Context) {
	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	d := c.MustGet("DB").(*db.DB)
	u := c.MustGet("User").(*models.User)

	node := &models.Node{}
	tags := []*models.Tag{}

	params := struct {
		Tags []string `json:"tags"`
	}{}

	if nid == 0 || err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NODE_NOT_FOUND})
		return
	}

	d.First(&node, "id = ?", nid)

	if node == nil || node.ID == 0 || !node.CanBeTaggedBy(u) {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NOT_ENOUGH_RIGHTS})
		return
	}

	err = c.BindJSON(&params)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.INCORRECT_DATA})
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
func (_ NodeController) PostLike(c *gin.Context) {
	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	d := c.MustGet("DB").(*db.DB)
	u := c.MustGet("User").(*models.User)

	node := &models.Node{}

	if nid == 0 || err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NODE_NOT_FOUND})
		return
	}

	d.First(&node, "id = ?", nid)

	if node == nil || node.ID == 0 || !node.CanBeLiked() {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NOT_ENOUGH_RIGHTS})
		return
	}

	isLiked := d.IsNodeLikedBy(node, u.ID)

	if isLiked {
		d.Model(&node).Association("Likes").Delete(u)
	} else {
		d.Model(&node).Association("Likes").Append(u)
	}

	c.JSON(http.StatusOK, gin.H{"is_liked": !isLiked})
}

// PostLock - POST /node/:id/lock - safely deletes node
func (_ NodeController) PostLock(c *gin.Context) {
	d := c.MustGet("DB").(*db.DB)
	u := c.MustGet("User").(*models.User)
	params := struct {
		IsLocked bool `json:"is_locked"`
	}{}

	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.INCORRECT_DATA})
		return
	}

	err = c.BindJSON(&params)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.INCORRECT_DATA})
		return
	}

	node := &models.Node{}

	if nid == 0 || err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NODE_NOT_FOUND})
		return
	}

	d.Unscoped().First(&node, "id = ?", nid)

	if node == nil || node.ID == 0 || !node.CanBeEditedBy(u) {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NOT_ENOUGH_RIGHTS})
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
func (_ NodeController) PostHeroic(c *gin.Context) {
	d := c.MustGet("DB").(*db.DB)
	u := c.MustGet("User").(*models.User)

	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.INCORRECT_DATA})
		return
	}

	node := &models.Node{}

	if nid == 0 || err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NODE_NOT_FOUND})
		return
	}

	d.First(&node, "id = ?", nid)

	if node == nil || node.ID == 0 || !node.CanBeHeroedBy(u) {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NOT_ENOUGH_RIGHTS})
		return
	}

	d.Model(&node).Update("is_heroic", !node.IsHeroic)

	c.JSON(http.StatusOK, gin.H{"is_heroic": node.IsHeroic})
}

// PostCellView - POST /node/:id/cell-view - sets cel display for node
func (_ NodeController) PostCellView(c *gin.Context) {
	d := c.MustGet("DB").(*db.DB)
	u := c.MustGet("User").(*models.User)
	params := struct {
		Flow models.NodeFlow `json:"flow"`
	}{}

	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.INCORRECT_DATA})
		return
	}

	err = c.BindJSON(&params)

	if err != nil || !models.NODE_FLOW_DISPLAY.Contains(params.Flow.Display) {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.INCORRECT_DATA})
		return
	}

	node := &models.Node{}

	if nid == 0 || err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NODE_NOT_FOUND})
		return
	}

	d.First(&node, "id = ?", nid)

	if node == nil || node.ID == 0 || !node.CanBeEditedBy(u) {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NOT_ENOUGH_RIGHTS})
		return
	}

	node.Flow = params.Flow

	d.Model(&node).Update("flow", node.Flow)

	c.JSON(http.StatusOK, gin.H{"flow": node.Flow})
}

// GetRelated - GET /node/:id/related - gets related nodes
func (_ NodeController) GetRelated(c *gin.Context) {
	d := c.MustGet("DB").(*db.DB)

	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.INCORRECT_DATA})
		return
	}

	if nid == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NODE_NOT_FOUND})
		return
	}

	node := &models.Node{}

	d.Preload("Tags").First(&node, "id = ?", nid)

	if node == nil || node.ID == 0 || node.DeletedAt != nil {
		c.JSON(http.StatusNotFound, gin.H{"related": &NodeRelatedResponse{}})
		return
	}

	if !node.IsFlowType() {
		c.JSON(http.StatusOK, gin.H{"related": &NodeRelatedResponse{}})
		return
	}

	if len(node.Tags) == 0 {
		c.JSON(http.StatusOK, gin.H{"related": &NodeRelatedResponse{}})
		return
	}

	var tagSimilarIds []uint
	var tagAlbumIds []uint

	for _, v := range node.Tags {
		if v.Title[:1] == "/" {
			tagAlbumIds = append(tagAlbumIds, v.ID)
		} else {
			tagSimilarIds = append(tagSimilarIds, v.ID)
		}
	}

	related := &NodeRelatedResponse{}

	var wg sync.WaitGroup
	wg.Add(2)

	albumsChan := make(chan map[string][]models.NodeRelatedItem)
	similarChan := make(chan []models.NodeRelatedItem)

	go d.GetNodeAlbumRelated(tagAlbumIds, []uint{node.ID}, node.Type, &wg, albumsChan)
	go d.GetNodeSimilarRelated(tagSimilarIds, []uint{node.ID}, node.Type, &wg, similarChan)

	wg.Wait()

	related.Albums = <-albumsChan
	related.Similar = <-similarChan

	c.JSON(http.StatusOK, gin.H{"related": related})
}

func (_ NodeController) PostNode(c *gin.Context) {
	d := c.MustGet("DB").(*db.DB)
	u := c.MustGet("User").(*models.User)

	if !u.CanCreateNode() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": codes.NOT_ENOUGH_RIGHTS})
		return
	}

	params := struct {
		Node models.Node `json:"node"`
	}{}

	err := c.BindJSON(&params)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.INCORRECT_DATA})
		return
	}

	node := &models.Node{}

	if params.Node.ID != 0 {
		d.First(&node, "id = ?", params.Node.ID)
		node.Cover = nil
		node.CoverID = 0
	} else {
		node.User = u
		node.UserID = u.ID
		node.Type = params.Node.Type
		node.IsPublic = true
		node.IsPromoted = true
	}

	if params.Node.Type == "" || !models.FLOW_NODE_TYPES.Contains(params.Node.Type) {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.INCORRECT_TYPE})
		return
	}

	if !node.CanBeEditedBy(u) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": codes.NOT_ENOUGH_RIGHTS})
		return
	}

	// Update previous node cover target to be null if its changed
	if node.Cover != nil &&
		node.CoverID != 0 &&
		(params.Node.Cover == nil || params.Node.Cover.ID == 0 || params.Node.Cover.ID != node.CoverID) {
		d.Model(&node.Files).Where("id IN (?)", node.CoverID).Update("target", nil)
	}

	// Validate node cover
	if params.Node.Cover != nil && params.Node.Cover.ID != 0 {
		if node.Cover == nil {
			node.Cover = &models.File{}
		}

		d.First(&models.File{}, "id = ?", params.Node.Cover.ID).Scan(&node.Cover)
		node.CoverID = params.Node.Cover.ID
	}

	// Finding out valid comment attaches and sorting them according to files_order
	originFiles := make([]uint, len(node.FilesOrder))
	copy(originFiles, node.FilesOrder)

	// Setting FilesOrder based on sorted Files array of input data
	params.Node.FilesOrder = make(models.CommaUintArray, 0)

	for _, v := range params.Node.Files {
		params.Node.FilesOrder = append(node.FilesOrder, v.ID)
	}

	if len(params.Node.FilesOrder) > 0 {
		ids, _ := params.Node.FilesOrder.Value()

		d.Order(gorm.Expr(fmt.Sprintf("FIELD(id, %s)", ids))).
			Find(&params.Node.Files, "id IN (?)", []uint(params.Node.FilesOrder))

		node.ApplyFiles(params.Node.Files)
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

	node.Title = params.Node.Title

	if len(node.Title) > 64 {
		node.Title = node.Title[:64]
	}

	node.ApplyBlocks(params.Node.Blocks)

	if val, ok := validation.NODE_VALIDATORS[node.Type]; ok {
		err = val(node)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// Update node metadata
	node.UpdateDescription()
	node.UpdateThumbnail()

	// Save node and its files
	d.Set("gorm:association_autoupdate", false).
		Set("gorm:association_save_reference", false).
		Save(&node).
		Association("Files").Replace(node.Files)

	// Node not saved
	if node.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.INCORRECT_DATA})
		return
	}

	// Updating current files target
	if len(node.FilesOrder) > 0 {
		d.Model(&node.Files).Where("id IN (?)", []uint(node.FilesOrder)).Update("Target", "node")
	}

	c.JSON(http.StatusOK, gin.H{"node": node})
}
