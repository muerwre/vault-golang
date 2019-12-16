package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type User struct {
	*gorm.Model

	Username    string `gorm:"unique;not null"`
	Password    string `json:"-"`
	Email       string
	Role        string
	Fullname    string
	Description string
	IsActivated string `json:"-"`
	Cover       File   `gorm:"foreignkey:CoverID"` // on delete null
	CoverID     uint   `gorm:"column:coverId"`
	Photo       File   `gorm:"foreignkey:CoverID"` // on delete null, eager
	PhotoID     uint   `gorm:"column:photoId"`
	Files       []File `json:"-" gorm:"foreignkey:userId"`
	// Nodes            []Node     `json:"-"`
	// Comments         []Comment  `json:"-"`
	// MessagesSent     []Message  `json:"messages_sent"`
	// MessagesReceived []Message  `json:"messages_received"`
	// Tags             []Tag      `json:"-"`
	// Likes            []Like     `json:"-"`
	// Tokens           []Token    `json:"-"`
	// NodeViews        []NodeView `json:"-"`
	// Socials          []Social   `json:"-"`
	LastSeen         time.Time
	LastSeenMessages time.Time
}
