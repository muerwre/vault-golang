package models

import (
	"github.com/muerwre/vault-golang/internal/service/notification/dto"
	"time"
)

type AppNotification struct {
	*Model

	App    string     `gorm:"column:app"`
	Type   string     `gorm:"column:type"`
	ItemID uint       `gorm:"column:item_id;index:item_id;"`
	SentAt *time.Time `gorm:"column:sent_at"`
}

func (AppNotification) TableName() string {
	return "app_notifications"
}

func (n AppNotification) ToDto() dto.NotificationDto {
	return dto.NotificationDto{
		CreatedAt: n.CreatedAt,
		ItemId:    n.ItemID,
		Type:      n.Type,
		Timestamp: n.CreatedAt,
	}
}
