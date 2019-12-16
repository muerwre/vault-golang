package models

import (
	"time"
)

type NodeView struct {
	ID uint

	User   User `json:"-"`
	UserID uint `gorm:"column:userId" json:"-"`

	Node   Node `json:"-"`
	NodeID uint `gorm:"column:nodeId" json:"-"`

	Visited time.Time
}

func (NodeView) TableName() string {
	return "node_view"
}
