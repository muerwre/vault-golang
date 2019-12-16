package models

import "time"

type MessageView struct {
	ID uint

	Dialog *User     `json:"-"`
	User   *User     `json:"-"`
	Viewed time.Time `json:"-"`
}

func (MessageView) TableName() string {
	return "message_view"
}
