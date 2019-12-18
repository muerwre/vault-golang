package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
)

type NodeController struct{}

var Node = &NodeController{}

func (a *NodeController) GetNode(c *gin.Context) {
	id := c.Param("id")
	uid := c.MustGet("UID").(uint)
	d := c.MustGet("DB").(*db.DB)

	node := &models.Node{}

	d.Preload("Tags").Preload("User").Preload("Cover").First(&node, "id = ?", id)

	if node.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.NODE_NOT_FOUND})
		return
	}

	files := make([]*models.File, len(node.FilesOrder))
	d.Where("id IN (?)", []string(node.FilesOrder)).Find(&files)
	node.Files = files

	if uid != 0 {
		node.IsLiked = d.IsNodeLikedBy(node, uid)
	}

	node.LikeCount = d.GetNodeLikeCount(node)

	c.JSON(http.StatusOK, gin.H{"node": node})
}
