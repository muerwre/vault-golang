package notify

import "time"

type NotifierItem struct {
	Type      string
	ItemId    uint
	CreatedAt time.Time
	Timestamp time.Time
}
