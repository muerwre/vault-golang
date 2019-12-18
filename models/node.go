package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type NodeBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
	Url  string `json:"url"`
}

type NodeBlocks []NodeBlock

type NodeFlow struct {
	*SimpleJson

	Display         string `json:"display"`
	ShowDescription bool   `json:"show_description"`
}

type Node struct {
	*Model

	Title       string           `json:"title"`
	Type        string           `json:"type"`
	IsPublic    bool             `json:"is_public"`
	IsPromoted  bool             `json:"is_promoted"`
	IsHeroic    bool             `json:"is_heroic"`
	Thumbnail   string           `json:"thumbnail"`
	Description string           `json:"description"`
	Blocks      NodeBlocks       `gorm:"type:longtext" json:"blocks"`
	Cover       *File            `gorm:"foreignkey:CoverID" json:"cover"` // on delete null
	CoverID     uint             `gorm:"column:coverId" json:"-"`
	User        *User            `json:"user" gorm:"foreignkey:UserID"`
	UserID      uint             `gorm:"column:userId" json:"-"`
	FilesOrder  CommaStringArray `gorm:"column:files_order;type:longtext;" json:"files_order"`
	Files       []*File          `gorm:"many2many:node_files_file;jointable_foreignkey:nodeId;association_jointable_foreignkey:fileId;" json:"files"`
	Tags        []*Tag           `gorm:"many2many:node_tags_tag;jointable_foreignkey:nodeId;association_jointable_foreignkey:tagId;" json:"tags"`
	Comments    []*Comment       `json:"-"`
	Likes       []*User          `gorm:"many2many:like;jointable_foreignkey:nodeId;association_jointable_foreignkey:userId;" json:"-"`
	Views       []*NodeView      `json:"-"`
	CommentedAt time.Time        `json:"commented_at" gorm:"column:commented_at"`
	Flow        NodeFlow         `json:"flow" gorm:"column:flow"`
	IsLiked     bool             `json:"is_liked" gorm:"-" sql:"-"`
	LikeCount   int              `json:"like_count" gorm:"-" sql:"-"`
}

type FlowNode struct {
	*Model
	Title       string    `json:"title"`
	Type        string    `json:"type"`
	Thumbnail   string    `json:"thumbnail"`
	Description string    `json:"description"`
	CommentedAt time.Time `json:"commented_at" gorm:"column:commented_at"`
	Flow        NodeFlow  `json:"flow" gorm:"column:flow"`
}

func (Node) TableName() string {
	return "node"
}

func (s *NodeBlocks) Scan(src interface{}) error {
	return json.Unmarshal(src.([]byte), &s)
}

func (s NodeBlocks) Value() (driver.Value, error) {
	fmt.Println("DECODER!", s)
	val, err := json.Marshal(s)
	return string(val), err
}

func (s *NodeFlow) Scan(src interface{}) error {
	return json.Unmarshal(src.([]byte), &s)
}

func (s NodeFlow) Value() (driver.Value, error) {
	val, err := json.Marshal(s)
	return string(val), err
}
