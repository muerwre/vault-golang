package models

import "time"

type Token struct {
	ID uint

	Token string `json:"-"`

	User   *User `json:"-" gorm:"foreignkey:UserID"`
	UserID uint  `gorm:"column:userId" json:"-"`

	CreatedAt time.Time `gorm:"column:created_at"`
}

func (Token) TableName() string {
	return "token"
}
