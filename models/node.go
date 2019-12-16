package models

import (
	"time"
)

type NodeBlock struct {
	*SimpleJson

	Type string `json:"type"`
	Text string `json:"text"`
	Url  string `json:"url"`
}

type NodeFlow struct {
	*SimpleJson

	Display         string `json:"display"`
	ShowDescription string `json:"show_description"`
}

type Node struct {
	*Model

	Title       string      `json:"title"`
	Type        string      `json:"type"`
	Blocks      []NodeBlock `gorm:"type:longtext" json:"blocks"`
	Cover       File        `gorm:"foreignkey:CoverID" json:"cover"` // on delete null
	CoverID     uint        `gorm:"column:coverId" json:"-"`
	User        User        `json:"user"`
	UserID      uint        `gorm:"column:userId" json:"-"`
	FilesOrder  StringArray `gorm:"column:files_order;type:longtext;" json:"files_order"`
	Files       []*File     `gorm:"many2many:node_files;" json:"files"`
	Tags        []*Tag      `gorm:"many2many:node_tags;" json:"tags"`
	Comments    []*Comment  `gorm:"many2many:node_comments;" json:"-"`
	Likes       []*User     `gorm:"many2many:node_likes;" json:"-"`
	IsPublic    bool        `json:"is_public"`
	IsPromoted  bool        `json:"is_promoted"`
	IsHeroic    bool        `json:"is_heroic"`
	Views       []NodeView  `json:"-"`
	CommentedAt *time.Time  `json:"commented_at"`
	Thumbnail   string      `json:"thumbnail"`
	Description string      `json:"description"`
}

func (Node) TableName() string {
	return "node"
}
