package dto

import "time"

type NotificationDto struct {
	Type      string
	ItemId    uint
	CreatedAt time.Time
	Timestamp time.Time
}
