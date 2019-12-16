package models

type Comment struct {
	*CommentLike

	User   *User `json:"user" gorm:"foreignkey:UserID"`
	UserID uint  `gorm:"column:userId" json:"-"`

	Node   *Node `json:"node" gorm:"foreignkey:NodeID"`
	NodeID uint  `gorm:"column:nodeId" json:"-"`
}

func (Comment) TableName() string {
	return "comment"
}
