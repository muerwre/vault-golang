package models

type Comment struct {
	*CommentLike

	User   *User `json:"user"`
	UserID uint  `gorm:"column:userId" json:"-"`

	Node   *Node `json:"node"`
	NodeID uint  `gorm:"column:nodeId" json:"-"`
}

func (Comment) TableName() string {
	return "comment"
}
