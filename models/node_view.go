package models

import (
	"time"
)

type NodeView struct {
	ID uint

	User   User `json:"-" gorm:"foreignkey:UserID"`
	UserID uint `gorm:"column:userId" json:"-"`

	Node   Node `json:"-" gorm:"foreignkey:NodeId"`
	NodeID uint `gorm:"column:nodeId" json:"-"`

	Visited time.Time
}

func (NodeView) TableName() string {
	return "node_view"
}
