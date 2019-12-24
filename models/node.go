package models

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/fatih/structs"
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

type FlowNodeTypes struct {
	IMAGE string
	VIDEO string
	TEXT  string
	AUDIO string
}

type ServiceNodeTypes struct {
	BORIS string
}

type NodeTypes struct {
	IMAGE string
	VIDEO string
	TEXT  string
	BORIS string
	AUDIO string
}

type NodeFlowDisplay struct {
	SINGLE     string
	VERTICAL   string
	HORIZONTAL string
	QUADRO     string
}

var FLOW_NODE_TYPES = FlowNodeTypes{
	IMAGE: "image",
	VIDEO: "video",
	TEXT:  "text",
	AUDIO: "audio",
}

var NODE_TYPES = NodeTypes{
	IMAGE: "image",
	VIDEO: "video",
	AUDIO: "audio",
	TEXT:  "text",
	BORIS: "boris",
}

var NODE_FLOW_DISPLAY = NodeFlowDisplay{
	SINGLE:     "single",
	VERTICAL:   "vertical",
	HORIZONTAL: "horizontal",
	QUADRO:     "quadro",
}

type NodeRelatedItem struct {
	Album     string `json:"-" sql:"album" gorm:"column:album"`
	Id        uint   `json:"id" sql:"id"`
	Thumbnail string `json:"thumbnail" sql:"thumbnail"`
	Title     string `json:"title" sql:"title"`
}

type Node struct {
	*Model

	Title       string         `json:"title"`
	Type        string         `json:"type"`
	IsPublic    bool           `json:"is_public"`
	IsPromoted  bool           `json:"is_promoted"`
	IsHeroic    bool           `json:"is_heroic"`
	Thumbnail   string         `json:"thumbnail"`
	Description string         `json:"description"`
	Blocks      NodeBlocks     `gorm:"type:longtext" json:"blocks"`
	Cover       *File          `gorm:"foreignkey:CoverID" json:"cover"` // on delete null
	CoverID     uint           `gorm:"column:coverId" json:"-"`
	User        *User          `json:"user" gorm:"foreignkey:UserID"`
	UserID      uint           `gorm:"column:userId" json:"-"`
	FilesOrder  CommaUintArray `gorm:"column:files_order;type:longtext;" json:"files_order"`
	Files       []*File        `gorm:"many2many:node_files_file;jointable_foreignkey:nodeId;association_jointable_foreignkey:fileId;" json:"files"`
	Tags        []*Tag         `gorm:"many2many:node_tags_tag;jointable_foreignkey:nodeId;association_jointable_foreignkey:tagId;" json:"tags"`
	Comments    []*Comment     `json:"-"`
	Likes       []*User        `gorm:"many2many:like;jointable_foreignkey:nodeId;association_jointable_foreignkey:userId;" json:"-"`
	Views       []*NodeView    `json:"-"`
	CommentedAt time.Time      `json:"commented_at" gorm:"column:commented_at"`
	Flow        NodeFlow       `json:"flow" gorm:"column:flow"`
	IsLiked     bool           `json:"is_liked" gorm:"-" sql:"-"`
	LikeCount   int            `json:"like_count" gorm:"-" sql:"-"`
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

func (f NodeFlowDisplay) Contains(t string) bool {
	for _, a := range structs.Map(f) {
		if a == t {
			return true
		}
	}

	return false
}

func (f FlowNodeTypes) Contains(t string) bool {
	for _, a := range structs.Map(f) {
		if a == t {
			return true
		}
	}

	return false
}

func (n Node) IsFlowType() bool {
	return FLOW_NODE_TYPES.Contains(n.Type)
}

func (n Node) CanBeCommented() bool {
	return n.Type == NODE_TYPES.BORIS || n.IsFlowType()
}

func (n Node) CanBeTaggedBy(user *User) bool {
	return n.IsFlowType() && (user.Role == USER_ROLES.ADMIN || n.UserID == user.ID)
}

func (n Node) CanBeEditedBy(user *User) bool {
	return n.IsFlowType() && (user.Role == USER_ROLES.ADMIN || n.UserID == user.ID)
}

func (n Node) CanBeLiked() bool {
	return n.IsFlowType()
}

func (n Node) CanBeHeroedBy(u *User) bool {
	return u.Role == USER_ROLES.ADMIN && n.IsFlowType()
}

func (n Node) CanHasFile(f *File) bool {
	switch n.Type {

	case NODE_TYPES.IMAGE:
		return f.Type == FILE_TYPES.IMAGE

	case NODE_TYPES.AUDIO:
		return f.Type == FILE_TYPES.AUDIO || f.Type == FILE_TYPES.IMAGE

	default:
		return false
	}
}
