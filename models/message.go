package models

type Message struct {
	*CommentLike

	From   *User `json:"from" gorm:"foreignkey:FromId"`
	FromID *User `gorm:"column:fromId" json:"-"`

	To   *User `json:"to" gorm:"foreignkey:ToID"`
	ToID *User `gorm:"column:toId" json:"-"`
}

func (Message) TableName() string {
	return "message"
}
