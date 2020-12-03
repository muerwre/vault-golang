package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/db/models"
	"github.com/muerwre/vault-golang/feature/node/constants"
	"github.com/muerwre/vault-golang/feature/node/request"
	"github.com/muerwre/vault-golang/feature/node/response"
	nodeUsecase "github.com/muerwre/vault-golang/feature/node/usecase"
	tagUsecase "github.com/muerwre/vault-golang/feature/tag/usecase"
	fileUsecase "github.com/muerwre/vault-golang/feature/upload/usecase"
	"github.com/muerwre/vault-golang/service/notification/controller"
	"github.com/muerwre/vault-golang/utils/codes"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type NodeController struct {
	node nodeUsecase.NodeUsecase
	file fileUsecase.FileUseCase
	tag  tagUsecase.TagUsecase
}

func (nc *NodeController) Init(db db.DB, config app.Config, notifier controller.NotificationService) *NodeController {
	nc.node = *new(nodeUsecase.NodeUsecase).Init(db, notifier)
	nc.file = *new(fileUsecase.FileUseCase).Init(db, config)
	nc.tag = *new(tagUsecase.TagUsecase).Init(db)

	return nc
}

// GetNode /node:id - returns single node with tags, likes count and files
func (nc *NodeController) GetNode(c *gin.Context) {
	u := c.MustGet("User").(*models.User)

	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NodeNotFound})
		return
	}

	node, err := nc.node.GetNodeWithLikesAndFiles(
		id,
		u.Role,
		u.ID,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NodeNotFound})
		return
	}

	node.Files = nc.file.UpdateFileMetadataIfNeeded(node.Files)

	c.JSON(http.StatusOK, gin.H{"node": node})
}

// GetNodeComments /node/:id/comments - returns comments for node
func (nc *NodeController) GetNodeComments(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NodeNotFound})
		return
	}

	take, _ := strconv.Atoi(c.Query("take"))
	skip, err := strconv.Atoi(c.Query("skip"))
	order := c.Query("order")

	comments, count, err := nc.node.GetComments(uint(id), take, skip, order)

	if err != nil {
		logrus.Warnf("Can't load comments for node: %+v", err)
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comments": comments, "comment_count": count})
}

// GetDiff /nodes/diff gets newer and older nodes
func (nc *NodeController) GetDiff(c *gin.Context) {
	// TODO: move to flow controller
	params := &request.NodeDiffParams{}
	err := c.Bind(&params)
	uid := c.MustGet("UID").(uint)

	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": codes.IncorrectData})
		return
	}

	params.Normalize()

	before, err := nc.node.GetDiffNodesBefore(params.Start)
	after, err := nc.node.GetDiffNodesAfter(params.End, params.Take)
	heroes, err := nc.node.GetDiffHeroes(params.WithHeroes)
	updated, exclude, err := nc.node.GetDiffUpdated(uid, params.WithUpdated)
	recent, err := nc.node.GetDiffRecent(exclude, params.WithRecent)
	valid, err := nc.node.GetDiffValid(params.Start, params.End, params.WithValid)

	resp := new(response.FlowResponse).Init(before, after, heroes, updated, recent, valid)

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

	comment, err := nc.node.GetDeletedComment(uint(cid), uint(nid), *u)

	if err != nil || *comment.NodeID != uint(nid) {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.CommentNotFound})
		return
	}

	if params.IsLocked {
		if err := nc.node.DeleteComment(comment); err != nil {
			logrus.Warnf("Unable to delete comment %d for node %d: %s", comment.ID, comment.NodeID, err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": codes.CantDeleteComment})
			return
		}

		if err = nc.node.PushCommentDeleteNotification(*comment); err != nil {
			logrus.Warnf(err.Error())
		}
	} else {
		if err := nc.node.RestoreComment(comment); err != nil {
			logrus.Warnf("Unable to restore comment %d for node %d: %s", comment.ID, comment.NodeID, err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": codes.CantRestoreComment})
			return
		}

		if err = nc.node.PushCommentRestoreNotification(*comment); err != nil {
			logrus.Warnf(err.Error())
		}
	}

	nc.node.UpdateNodeCommentedAt(uint(nid))

	c.JSON(http.StatusOK, gin.H{"deteled_at": &comment.DeletedAt})
}

// PostComment - /node/:id/comment saving and creating comments
func (nc *NodeController) PostComment(c *gin.Context) {
	u := c.MustGet("User").(*models.User)

	data := &models.Comment{}
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if nid == 0 || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.NodeNotFound})
		return
	}

	node, err := nc.node.GetCommentableNodeById(uint(nid))
	if err != nil || !node.CanBeCommented() {
		logrus.Warnf(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.NodeNotFound})
		return
	}

	comment, err := nc.node.LoadCommentFromData(data.ID, node, u)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lostFiles, err := nc.node.UpdateCommentFiles(data, comment)
	if err != nil {
		logrus.Warnf("Can't load node files while updating comment %d for node %d: %s", node.ID, comment.ID, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.CantSaveComment})
		return
	}

	if err := nc.node.ValidateAndUpdateCommentText(data, comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err = nc.node.SaveCommentWithFiles(comment); err != nil {
		logrus.Warnf("Failed to save comment %d for node: %s", comment.ID, node.ID, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.CantSaveComment})
		return
	}

	nc.node.UnsetFilesTarget(lostFiles)
	nc.node.UpdateBriefFromComment(node, comment)
	nc.node.UpdateFilesMetadata(data.Files, comment.Files)
	nc.node.UpdateNodeSeen(uint(nid), u.ID)
	nc.node.UpdateNodeCommentedAt(uint(nid))

	c.JSON(http.StatusOK, gin.H{"comment": comment})

	if err = nc.node.PushCommentCreateNotificationIfNeeded(*data, *comment); err != nil {
		logrus.Warnf(err.Error())
	}
}

// PostTags - POST /node/:id/tags - updates node tags
func (nc *NodeController) PostTags(c *gin.Context) {
	u := c.MustGet("User").(*models.User)

	params := request.NodeTagsPostRequest{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if nid == 0 || err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NodeNotFound})
		return
	}

	node, err := nc.node.GetTaggableNodeById(uint(nid), u)
	if err != nil {
		logrus.Warnf("Node %d not found: %s", nid, err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NodeNotFound})
		return
	}

	tags, err := nc.tag.FindOrCreateTags(params.Tags)
	if err != nil {
		logrus.Warnf("Can't find or create tags for node %d: %s", nid, err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": codes.CantSaveNode})
		return
	}

	if err := nc.node.UpdateNodeTags(node, tags); err != nil {
		logrus.Warnf("Can't save node tags for node %d: %s\n tags: %+v\n", nid, err.Error(), tags)
		c.JSON(http.StatusNotFound, gin.H{"error": codes.CantSaveNode})
		return
	}

	c.JSON(http.StatusOK, gin.H{"node": node})
}

// PostLike - POST /node/:id/like - likes or dislikes node
func (nc NodeController) PostLike(c *gin.Context) {
	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	u := c.MustGet("User").(*models.User)

	node, err := nc.node.GetNodeWithLikesAndFiles(int(nid), u.Role, u.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NodeNotFound})
		return
	}

	if err := nc.node.UpdateNodeLikeByUser(node, u, !node.IsLiked); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": codes.CantSaveNode})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_liked": !node.IsLiked})
}

// LockNode - POST /node/:id/lock - safely deletes node
func (nc NodeController) LockNode(c *gin.Context) {
	u := c.MustGet("User").(*models.User)
	params := request.NodeLockRequest{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	node, err := nc.node.GetDeletedNode(uint(nid), u)
	if err != nil {
		logrus.Warnf("Can't get deleted node %d: %s", nid, err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NodeNotFound})
		return
	}

	if err := nc.node.UpdateNodeLocked(node, params.IsLocked); err != nil {
		logrus.Warnf("Can't lock node %d: %s", nid, err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": codes.CantSaveNode})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted_at": node.DeletedAt})
}

// PostHeroic - POST /node/:id/heroic - sets heroic status to node
func (nc NodeController) PostHeroic(c *gin.Context) {
	u := c.MustGet("User").(*models.User)

	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || nid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	node, err := nc.node.GetHeroeableNodeById(uint(nid), u)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.NodeNotFound})
		return
	}

	if err := nc.node.UpdateNodeIsHeroic(node, !node.IsHeroic); err != nil {
		logrus.Warnf("Can't update node isHeroic for node %d: %s", node.ID, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.CantSaveNode})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_heroic": node.IsHeroic})
}

// PostCellView - POST /node/:id/cell-view - sets cel display for node
func (nc NodeController) PostCellView(c *gin.Context) {
	u := c.MustGet("User").(*models.User)
	params := request.NodeCellViewPostRequest{}
	if err := c.BindJSON(&params); err != nil || !constants.NODE_FLOW_DISPLAY.Contains(params.Flow.Display) {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || nid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	node, err := nc.node.GetEditableNodeById(uint(nid), u)
	if err != nil {
		logrus.Warnf(err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NodeNotFound})
		return
	}

	if err := nc.node.UpdateNodeFlow(node, params.Flow); err != nil {
		logrus.Warnf("Can't update flow settings for node %d: %+v\nFlow: %+v", nid, err.Error(), params.Flow)
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.CantSaveNode})
		return
	}

	c.JSON(http.StatusOK, gin.H{"flow": node.Flow})
}

// GetRelated - GET /node/:id/related - gets related nodes
func (nc NodeController) GetRelated(c *gin.Context) {
	nid, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if err != nil || nid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	related, err := nc.node.GetNodeRelated(uint(nid))

	c.JSON(http.StatusOK, gin.H{"related": related})
}

func (nc NodeController) PostNode(c *gin.Context) {
	user := c.MustGet("User").(*models.User)

	if !user.CanCreateNode() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": codes.NotEnoughRights})
		return
	}

	data := request.NodePostRequest{}

	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	n := data.ToNode()

	node, err := nc.node.LoadNodeFromData(*n, user)

	if err != nil {
		logrus.Warnf("Can't load node from data: %s\nData:\n%+v", err.Error(), data)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Update previous node cover target to be null if its changed
	if err = nc.node.UpdateNodeCoverIfChanged(*n, node); err != nil {
		logrus.Warnf("Can't load node cover from data: %s\nData:\n%+v", err.Error(), data)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	lostFiles, err := nc.node.UpdateNodeFiles(*n, node)

	if err != nil {
		logrus.Warnf("Can't load node files while updating node %d: %s", n.ID, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nc.node.UpdateNodeTitle(*n, node)
	nc.node.UpdateNodeVisibility(*n, node)

	if err = nc.node.UpdateNodeBlocks(*n, node); err != nil {
		logrus.Warnf("Received suspicious blocks while updating node %d: %s", n.ID, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	node.UpdateDescription()
	node.UpdateThumbnail()

	if err = nc.node.SaveNodeWithFiles(node); err != nil {
		logrus.Warnf("Failed to save node %d: %s", node.ID, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	nc.node.UpdateFilesMetadata(data.Files, node.Files)
	nc.node.SetFilesTarget(node.FilesOrder, "node")
	nc.node.UnsetFilesTarget(lostFiles)
	nc.node.UnsetNodeCoverTarget(*n, node)

	if err = nc.node.PushNodeCreateNotificationIfNeeded(*n, *node); err != nil {
		logrus.Warnf(err.Error())
	}

	c.JSON(http.StatusOK, gin.H{"node": node})
}
