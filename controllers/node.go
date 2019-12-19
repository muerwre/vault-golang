package controllers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
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

var FlowNodeTypes = []string{"image", "video", "text"}
var FlowNodeCriteria = "is_promoted = 1 AND is_public = 1 AND type IN (?)"

type NodeController struct{}

var Node = &NodeController{}

// GetNode /node:id - returns single node with tags, likes count and files
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

// GetNodeComments /node/:id/comments - returns comments for node
func (a *NodeController) GetNodeComments(c *gin.Context) {
	id := c.Param("id")
	d := c.MustGet("DB").(*db.DB)

	comments := &[]*models.Comment{}

	d.Preload("User").Preload("Files").Preload("User.Photo").Where("nodeId = ?", id).Order("created_at").Find(&comments)

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

	q := d.Model(&models.Node{}).Where(FlowNodeCriteria, FlowNodeTypes)

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

	c.BindJSON(&params)

	comment := &models.Comment{}
	d.Unscoped().Where("id = ?", cid).First(&comment)

	if comment == nil || comment.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.COMMENT_NOT_FOUND})
	}

	if !u.CanEditComment(comment) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": codes.NOT_ENOUGH_RIGHTS})
	}

	if params.IsLocked {
		d.Delete(&comment)
	} else {
		comment.DeletedAt = nil
		d.Unscoped().Update(&comment)
	}

	c.JSON(http.StatusOK, gin.H{"deteled_at": &comment.DeletedAt})
}
