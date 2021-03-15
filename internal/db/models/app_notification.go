package models

import "time"

type AppNotification struct {
	*Model

	App    string    `gorm:"column:app"`
	Type   string    `gorm:"column:type"`
	ItemID uint      `gorm:"column:itemId;index:item_id;"`
	SentAt time.Time `gorm:"column:sent_at"`
	UserID uint      `gorm:"column:userId" json:"-"`
}

func (AppNotification) TableName() string {
	return "app_notifications"
}
