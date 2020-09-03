package models

type Notification struct {
	*Model

	Type   string `gorm:"column:type"`
	ItemID uint   `gorm:"column:item_id;index:item_id;"`
}
