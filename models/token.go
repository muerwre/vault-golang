package models

import (
	"time"

	"github.com/muerwre/vault-golang/utils/passwords"
)

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

func (t *Token) New(username string) {
	t.Token, _ = passwords.HashPassword(username + string(time.Now().String()))
}
