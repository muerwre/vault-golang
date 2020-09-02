package models

type NodeWatch struct {
	*Model

	User   User `json:"-" gorm:"foreignkey:UserID"`
	UserID uint `gorm:"column:userId;uniqueIndex:user_node,unique;" json:"-"`

	Node   Node `json:"-" gorm:"foreignkey:NodeId;"`
	NodeID uint `gorm:"column:nodeId;uniqueIndex:user_node;" json:"-"`
}

func (NodeWatch) TableName() string {
	return "node_watch"
}
