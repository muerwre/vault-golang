package models

import "time"

type Notification struct {
	*Model

	Type   string    `gorm:"column:type"`
	ItemID uint      `gorm:"column:itemId;index:item_id;"`
	Time   time.Time `gorm:"column:time"`
	UserID uint      `gorm:"column:userId" json:"-"`
}

const (
	NotificationTypeNode    string = "node"
	NotificationTypeComment string = "comment"
)
