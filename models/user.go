package models

import (
	"time"
)

type User struct {
	*Model

	Username         string     `gorm:"unique;not null" json:"username"`
	Password         string     `json:"-"`
	Email            string     `json:"email"`
	Role             string     `json:"role"`
	Fullname         string     `json:"fullname"`
	Description      string     `json:"description"`
	IsActivated      string     `json:"-"`
	Cover            File       `gorm:"foreignkey:CoverID;" json:"cover"` // on delete null
	CoverID          uint       `gorm:"column:coverId" json:"-"`
	Photo            File       `gorm:"foreignkey:PhotoID;" json:"photo"` // on delete null, eager
	PhotoID          uint       `gorm:"column:photoId" json:"-"`
	Files            []File     `gorm:"foreignkey:userId" json:"-"`
	Nodes            []Node     `json:"-"`
	Comments         []Comment  `json:"-"`
	MessagesSent     []Message  `json:"-"`
	MessagesReceived []Message  `json:"-"`
	Tags             []Tag      `json:"-"`
	Likes            []Node     `gorm:"many2many:node_likes;" json:"-"`
	Tokens           []Token    `json:"-"`
	NodeViews        []NodeView `json:"-"`
	Socials          []Social   `json:"-"`
	LastSeen         time.Time  `json:"last_seen"`
	LastSeenMessages time.Time  `json:"last_seen_messages"`
}

func (User) TableName() string {
	return "user"
}
