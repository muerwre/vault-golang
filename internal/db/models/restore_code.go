package models

import "time"

type RestoreCode struct {
	ID uint

	Code   string `json:"code"`
	User   *User  `json:"user" gorm:"foreignkey:UserID"`
	UserID uint   `gorm:"column:userId" json:"-"`

	CreatedAt time.Time
}

func (RestoreCode) TableName() string {
	return "restore_code"
}
