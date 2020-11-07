package request

import (
	"github.com/muerwre/vault-golang/models"
	"time"
)

type NodeDiffParams struct {
	Start       *time.Time `json:"start" form:"start"`
	End         *time.Time `json:"end" form:"end"`
	Take        uint       `json:"take" form:"take"`
	WithHeroes  bool       `json:"with_heroes" form:"with_heroes"`
	WithUpdated bool       `json:"with_updated" form:"with_updated"`
	WithRecent  bool       `json:"with_recent" form:"with_recent"`
	WithValid   bool       `json:"with_valid" form:"with_valid"`
}

func (p *NodeDiffParams) Normalize() {
	if p.Take == 0 {
		p.Take = 40
	}

	if p.Start == nil {
		now := time.Now()
		p.Start = &now
	}

	if p.End == nil {
		now := time.Now().Add(-30 * 24 * time.Hour)
		p.End = &now
	}
}

type NodeCellViewPostRequest struct {
	Flow models.NodeFlow `json:"flow"`
}

type NodePostRequest struct {
	Node models.Node `json:"node"`
}

type NodeLockCommentRequest struct {
	IsLocked bool `json:"is_locked"`
}

type NodeTagsPostRequest struct {
	Tags []string `json:"tags"`
}

type NodeLockRequest struct {
	IsLocked bool `json:"is_locked"`
}
