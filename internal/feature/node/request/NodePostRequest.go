package request

import (
	"github.com/muerwre/vault-golang/db/models"
	"time"
)

type NodePostRequest struct {
	ID          uint                  `gorm:"primary_key" json:"id"`
	Title       string                `json:"title"`
	Type        string                `json:"type"`
	IsPublic    bool                  `json:"is_public"`
	IsPromoted  bool                  `json:"is_promoted"`
	IsHeroic    bool                  `json:"is_heroic"`
	Thumbnail   string                `json:"thumbnail"`
	Description string                `json:"description"`
	Blocks      models.NodeBlocks     `json:"blocks"`
	Cover       *models.File          `json:"cover"` // on delete null
	User        *models.User          `json:"user" gorm:"foreignkey:UserID"`
	FilesOrder  models.CommaUintArray `json:"files_order"`
	Files       []*models.File        `json:"files"`
	Tags        []*models.Tag         `json:"tags"`
	CommentedAt time.Time             `json:"commented_at" gorm:"column:commented_at"`
	Flow        models.NodeFlow       `json:"flow"`
}

func (r *NodePostRequest) ToNode() *models.Node {
	n := &models.Node{
		Model: &models.Model{
			ID: r.ID,
		},
	}

	n.Title = r.Title
	n.Type = r.Type
	n.IsPublic = r.IsPublic
	n.IsPromoted = r.IsPromoted
	n.IsHeroic = r.IsHeroic
	n.Thumbnail = r.Thumbnail
	n.Description = r.Description
	n.Blocks = r.Blocks
	n.Cover = r.Cover
	n.User = r.User
	n.FilesOrder = r.FilesOrder
	n.Files = r.Files
	n.Tags = r.Tags
	n.CommentedAt = r.CommentedAt
	n.Flow = r.Flow

	return n
}
