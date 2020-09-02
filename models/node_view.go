package models

import (
	"time"
)

type NodeView struct {
	ID uint

	User   User `json:"-" gorm:"foreignkey:UserID"`
	UserID uint `gorm:"column:userId;uniqueIndex:user_node;" json:"-"`

	Node   Node `json:"-" gorm:"foreignkey:NodeId"`
	NodeID uint `gorm:"column:nodeId;uniqueIndex:user_node;" json:"-"`

	Visited time.Time
}

func (NodeView) TableName() string {
	return "node_view"
}
