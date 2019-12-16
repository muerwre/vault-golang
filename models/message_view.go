package models

import "time"

type MessageView struct {
	ID uint

	Dialog   *User     `json:"-" gorm:"foreignkey:DialogID"`
	DialogId uint      `gorm:"column:dialogId"`
	User     *User     `json:"-" gorm:"foreignkey:UserID"`
	UserId   uint      `gorm:"column:userId" json:"-"`
	Viewed   time.Time `json:"-"`
}

func (MessageView) TableName() string {
	return "message_view"
}
