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

// GetNode /node:id returns single node
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

	files_chan := make(chan []*models.File)

	go func() {
		files := make([]*models.File, len(node.FilesOrder))
		d.Where("id IN (?)", []string(node.FilesOrder)).Find(&files)
		files_chan <- files
	}()

	if uid != 0 {
		node.IsLiked = d.IsNodeLikedBy(node, uid)
	}

	node.LikeCount = d.GetNodeLikeCount(node)
	node.Files = <-files_chan

	c.JSON(http.StatusOK, gin.H{"node": node})
}

func (a *NodeController) GetNodeComments(c *gin.Context) {
	id := c.Param("id")
	d := c.MustGet("DB").(*db.DB)

	comments := &[]*models.Comment{}

	d.Preload("User").Preload("Files").Where("nodeId = ?", id).Order("created_at").Find(&comments)

	for _, v := range *comments {
		v.SortFiles()
	}

	c.JSON(http.StatusAccepted, gin.H{"comments": comments})
}
