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

	Title       string `json:"title"`
	Type        string `json:"type"`
	IsPublic    bool   `json:"is_public"`
	IsPromoted  bool   `json:"is_promoted"`
	IsHeroic    bool   `json:"is_heroic"`
	Thumbnail   string `json:"thumbnail"`
	Description string `json:"description"`

	Blocks []NodeBlock `gorm:"type:longtext" json:"blocks"`

	Cover   File `gorm:"foreignkey:CoverID" json:"cover"` // on delete null
	CoverID uint `gorm:"column:coverId" json:"-"`

	User   User `json:"user" gorm:"foreignkey:UserID"`
	UserID uint `gorm:"column:userId" json:"-"`

	FilesOrder  StringArray `gorm:"column:files_order;type:longtext;" json:"files_order"`
	Files       []*File     `gorm:"many2many:node_files_file;association_jointable_foreignkey:nodeId;" json:"files"`
	Tags        []*Tag      `gorm:"many2many:node_tags_tag;association_jointable_foreignkey:nodeId;" json:"tags"`
	Comments    []*Comment  `json:"-"`
	Likes       []*User     `gorm:"many2many:like;association_jointable_foreignkey:nodeId;" json:"-"`
	Views       []*NodeView `json:"-"`
	CommentedAt *time.Time  `json:"commented_at"`
}

func (Node) TableName() string {
	return "node"
}
