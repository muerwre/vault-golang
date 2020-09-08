package controllers

import (
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/controllers/usecase"
	"github.com/muerwre/vault-golang/request"
	"github.com/muerwre/vault-golang/response"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
)

type NodeController struct {
	db          db.DB
	nodeUsecase usecase.NodeUsecase
	fileUsecase usecase.FileUseCase
}

func (nc *NodeController) Init(db db.DB, config app.Config) *NodeController {
	nc.nodeUsecase = *new(usecase.NodeUsecase).Init(db)
	nc.fileUsecase = *new(usecase.FileUseCase).Init(db, config)

	nc.db = db
	return nc
}

// GetNode /node:id - returns single node with tags, likes count and files
func (nc *NodeController) GetNode(c *gin.Context) {
	uid := c.MustGet("UID").(uint)
	u := c.MustGet("User").(*models.User)
	d := nc.db

	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NodeNotFound})
		return
	}

	node, err := nc.db.NodeRepository.GetFullNode(
		id,
		u.Role == models.USER_ROLES.ADMIN,
		u.ID,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NodeNotFound})
		return
	}

	if uid != 0 {
		node.IsLiked = d.NodeRepository.IsNodeLikedBy(node, uid)

		nc.db.NodeViewRepository.UpdateView(uid, node.ID)
	}

	node.LikeCount = d.NodeRepository.GetNodeLikeCount(node)
	node.Files, _ = nc.db.FileRepository.GetFilesByIds([]uint(node.FilesOrder))

	node.SortFiles()
	node.Files = nc.fileUsecase.UpdateFileMetadataIfNeeded(node.Files)

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

	comments, count := nc.db.NodeRepository.GetComments(id, take, skip, order)

	c.JSON(http.StatusOK, gin.H{"comments": comments, "comment_count": count})
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

	q := nc.db.Preload("User").Preload("User.Photo").Model(&models.Node{})
	// TODO: move to repo
	q = nc.db.NodeRepository.WhereIsFlowNode(q)

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
			q.Order("commented_at DESC, created_at DESC").
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
	u := c.MustGet("User").(*models.User)
	params := request.NodeLockCommentRequest{}

	cid, err := strconv.ParseUint(c.Param("cid"), 10, 32)

	if cid == 0 || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.NodeNotFound})
		return
	}

	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if nid == 0 || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.NodeNotFound})
		return
	}

	err = c.BindJSON(&params)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	comment, err := nc.db.NodeRepository.GetCommentByIdWithDeleted(uint(cid))

	if err != nil || *comment.NodeID != uint(nid) {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.CommentNotFound})
		return
	}

	if !u.CanEditComment(comment) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": codes.NotEnoughRights})
		return
	}

	if params.IsLocked {
		if err := nc.nodeUsecase.DeleteComment(comment); err != nil {
			logrus.Warnf("Unable to delete comment %d for node %d: %s", comment.ID, comment.NodeID, err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": codes.CantDeleteComment})
			return
		}
	} else {
		if err := nc.nodeUsecase.RestoreComment(comment); err != nil {
			logrus.Warnf("Unable to restore comment %d for node %d: %s", comment.ID, comment.NodeID, err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": codes.CantRestoreComment})
			return
		}
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

	node, err := nc.db.NodeRepository.GetById(uint(nid))

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

	comment, err := nc.nodeUsecase.LoadCommentFromData(data.ID, node, u)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lostFiles, err := nc.nodeUsecase.UpdateCommentFiles(data, comment)

	if err != nil {
		logrus.Warnf("Can't load node files while updating comment %d for node %d: %s", node.ID, comment.ID, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.CantSaveComment})
		return
	}

	if err := nc.nodeUsecase.UpdateCommentText(data, comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err = nc.db.NodeRepository.SaveCommentWithFiles(comment); err != nil {
		logrus.Warnf("Failed to save comment %d for node: %s", comment.ID, node.ID, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.CantSaveComment})
		return
	}

	nc.nodeUsecase.UnsetFilesTarget(lostFiles)
	nc.nodeUsecase.UpdateBriefFromComment(node, comment)
	nc.nodeUsecase.UpdateFilesMetadata(data.Files, comment.Files)
	nc.nodeUsecase.UpdateNodeCommentedAt(uint(nid), &comment.CreatedAt)
	nc.nodeUsecase.UpdateNodeSeen(uint(nid), u.ID)

	c.JSON(http.StatusOK, gin.H{"comment": comment})
}

// PostTags - POST /node/:id/tags - updates node tags
func (nc *NodeController) PostTags(c *gin.Context) {
	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	u := c.MustGet("User").(*models.User)

	tags := []*models.Tag{}

	params := request.NodeTagsPostRequest{}

	if nid == 0 || err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NodeNotFound})
		return
	}

	node, err := nc.db.NodeRepository.GetById(uint(nid))

	if err != nil {
		logrus.Warnf("Node %d not found: %s", nid, err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NodeNotFound})
		return
	}

	if !node.CanBeTaggedBy(u) {
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

	if len(params.Tags) == 0 {
		c.JSON(http.StatusOK, gin.H{"node": node})
		return
	}

	nc.db.Where("title IN (?)", params.Tags).Find(&tags)

	for _, v := range params.Tags {
		if !models.TagArrayContains(tags, v) && len(v) > 0 {
			tag := models.Tag{Title: v}
			nc.db.Set("gorm:association_autoupdate", false).Save(&tag)
			tags = append(tags, &tag)
		}
	}

	nc.db.Model(&node).Association("Tags").Replace(tags)

	c.JSON(http.StatusOK, gin.H{"node": node})
}

// PostLike - POST /node/:id/like - likes or dislikes node
func (nc NodeController) PostLike(c *gin.Context) {
	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	d := nc.db
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
	d := nc.db
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
	d := nc.db
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
	d := nc.db
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

	related, err := nc.nodeUsecase.GetNodeRelated(uint(nid))

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

	node, err := nc.nodeUsecase.LoadNodeFromData(data, user)

	if err != nil {
		logrus.Warnf("Can't load node from data: %s\nData:\n%+v", err.Error(), data)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Update previous node cover target to be null if its changed
	if err = nc.nodeUsecase.UpdateNodeCoverIfChanged(data, node); err != nil {
		logrus.Warnf("Can't load node cover from data: %s\nData:\n%+v", err.Error(), data)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	lostFiles, err := nc.nodeUsecase.UpdateNodeFiles(data, node)

	if err != nil {
		logrus.Warnf("Can't load node files while updating node %d: %s", data.ID, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nc.nodeUsecase.UpdateNodeTitle(data, node)

	if err = nc.nodeUsecase.UpdateNodeBlocks(data, node); err != nil {
		logrus.Warnf("Received suspicious blocks while updating node %d: %s", data.ID, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	node.UpdateDescription()
	node.UpdateThumbnail()

	if err = nc.db.NodeRepository.SaveNodeWithFiles(node); err != nil {
		logrus.Warnf("Failed to save node %d: %s", node.ID, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	nc.nodeUsecase.UpdateFilesMetadata(data.Files, node.Files)
	nc.nodeUsecase.SetFilesTarget(node.FilesOrder, "node")
	nc.nodeUsecase.UnsetFilesTarget(lostFiles)
	nc.nodeUsecase.UnsetNodeCoverTarget(data, node)

	c.JSON(http.StatusOK, gin.H{"node": node})
}
