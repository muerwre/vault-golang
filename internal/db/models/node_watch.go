package models

type NodeWatch struct {
	*Model

	User   User `json:"-" gorm:"foreignkey:UserID"`
	UserID uint `json:"-" gorm:"column:userId;uniqueIndex:user_node;"`

	Node   Node `json:"-" gorm:"foreignkey:NodeId;"`
	NodeID uint `json:"-" gorm:"column:nodeId;uniqueIndex:user_node;"`

	Active bool `json:"-" gorm:"column:active"`
}

func (NodeWatch) TableName() string {
	return "node_watch"
}
