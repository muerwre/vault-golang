package models

import (
	"time"

	"github.com/muerwre/vault-golang/utils/passwords"
)

type User struct {
	*Model

	Username         string     `gorm:"unique;not null" json:"username"`
	Password         string     `json:"-"`
	Email            string     `json:"-"`
	Role             string     `json:"role"`
	Fullname         string     `json:"fullname"`
	Description      string     `json:"description"`
	IsActivated      string     `json:"-"`
	Cover            *File      `gorm:"foreignkey:CoverID;" json:"cover"` // on delete null
	CoverID          uint       `gorm:"column:coverId" json:"-"`
	Photo            *File      `gorm:"foreignkey:PhotoID;" json:"photo"` // on delete null, eager
	PhotoID          uint       `gorm:"column:photoId" json:"-"`
	Files            []File     `gorm:"foreignkey:userId" json:"-"`
	Nodes            []Node     `json:"-"`
	Comments         []Comment  `json:"-"`
	MessagesSent     []Message  `json:"-"`
	MessagesReceived []Message  `json:"-"`
	Tags             []Tag      `json:"-"`
	Likes            []Node     `gorm:"many2many:like;association_jointable_foreignkey:userId;" json:"-"`
	Tokens           []Token    `json:"-"`
	NodeViews        []NodeView `json:"-"`
	Socials          []Social   `json:"-"`
	LastSeen         time.Time  `json:"last_seen"`
	LastSeenMessages time.Time  `json:"last_seen_messages"`

	NewPassword string `json:"-" gorm:"-" sql:"-"`
}

func (User) TableName() string {
	return "user"
}

func (u *User) IsValidPassword(password string) bool {
	return password != "" && u.Password != "" && passwords.CheckPasswordHash(password, u.Password)
}

func (u User) CanEditComment(c *Comment) bool {
	return c.UserID != 0 && (u.ID == c.UserID || u.Role == "admin")
}
