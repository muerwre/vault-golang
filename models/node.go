package models

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/muerwre/vault-golang/constants"
	"time"

	"github.com/fatih/structs"
	"github.com/muerwre/vault-golang/utils"
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

const (
	BlockTypeText  string = "text"
	BlockTypeVideo string = "video"
)

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
	CoverID     *uint          `gorm:"column:coverId" json:"-"`
	User        *User          `json:"user" gorm:"foreignkey:UserID"`
	UserID      *uint          `gorm:"column:userId" json:"-"`
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
	return n.IsFlowType() && (user.Role == USER_ROLES.ADMIN || *n.UserID == user.ID)
}

func (n Node) CanBeEditedBy(user *User) bool {
	return n.IsFlowType() && (user.Role == USER_ROLES.ADMIN || *n.UserID == user.ID)
}

func (n Node) CanBeLiked() bool {
	return n.IsFlowType()
}

// CanBeHeroedBy - checks if node can be set as promoted to front slider by this user
func (n Node) CanBeHeroedBy(u *User) bool {
	return u.Role == USER_ROLES.ADMIN && n.IsFlowType()
}

// CanHasFile - checks if node can has file of type
func (n Node) CanHasFile(f File) bool {
	switch n.Type {
	case NODE_TYPES.IMAGE:
		return f.Type == constants.FileTypeImage

	case NODE_TYPES.AUDIO:
		return f.Type == constants.FileTypeAudio || f.Type == constants.FileTypeImage

	default:
		return false
	}
}

// CanHasBlock - checks if node can has block of type
func (n Node) CanHasBlock(b NodeBlock) bool {
	switch n.Type {
	case NODE_TYPES.TEXT:
		return b.Type == BlockTypeText
	case NODE_TYPES.VIDEO:
		return b.Type == BlockTypeVideo
	default:
		return false
	}
}

// ApplyFiles - sets node files with validation
func (n *Node) ApplyFiles(files []*File) {
	n.Files = make([]*File, 0)
	n.FilesOrder = make(CommaUintArray, 0)

	for i := 0; i < len(files); i += 1 { // TODO: limit files count
		if n.CanHasFile(*files[i]) {
			n.Files = append(n.Files, files[i])
			n.FilesOrder = append(n.FilesOrder, files[i].ID)
		}
	}
}

// ApplyBlocks - sets node blocks with validation
func (n *Node) ApplyBlocks(blocks []NodeBlock) {
	n.Blocks = make([]NodeBlock, 0)

	for _, v := range blocks {
		if n.CanHasBlock(v) && v.IsValid() {
			n.Blocks = append(n.Blocks, v)
		}
	}
}

// IsValid - validates node block
func (b NodeBlock) IsValid() bool {
	return (b.Type == BlockTypeText && len(b.Text) > 0) ||
		(b.Type == BlockTypeVideo && len(b.Url) > 0 && utils.GetThumbFromUrl(b.Url) != "")
}

// FirstBlockOfType - gets block file of type (t)
func (n Node) FirstBlockOfType(t string) int {
	for k, v := range n.Blocks {
		if v.Type == t {
			return k
		}
	}

	return -1
}

// FirstFileOfType - gets first file of type (t)
func (n Node) FirstFileOfType(t string) int {
	for k, v := range n.Files {
		if v.Type == t {
			return k
		}
	}

	return -1
}

// UpdateDescription - generates node brief description from node's body
func (n *Node) UpdateDescription() {
	if n.Type == NODE_TYPES.TEXT {
		textBlock := n.Blocks[n.FirstBlockOfType(BlockTypeText)]

		if len(textBlock.Text) > 64 {
			n.Description = textBlock.Text
			return
		}
	}
}

// UpdateDescription - generates node thumbnail image from node's body
func (n *Node) UpdateThumbnail() {
	if n.Type == NODE_TYPES.IMAGE || n.Type == NODE_TYPES.AUDIO {
		i := n.FirstFileOfType(constants.FileTypeImage)

		if i >= 0 {
			n.Thumbnail = n.Files[i].Url
			return
		}
	}

	if n.Type == NODE_TYPES.VIDEO {
		i := n.FirstBlockOfType(BlockTypeVideo)

		if url := utils.GetThumbFromUrl(n.Blocks[i].Url); url != "" {
			n.Thumbnail = url

			return
		}
	}
}

// SortFiles - sorts comment files according to files_order
func (n *Node) SortFiles() {
	if len(n.FilesOrder) == 0 || len(n.Files) == 0 {
		return
	}

	filesWithIds := make(map[uint]*File, len(n.FilesOrder))
	files := make([]*File, len(n.FilesOrder))

	for i := 0; i < len(n.Files); i += 1 {
		k := n.Files[i]
		filesWithIds[k.ID] = k
	}

	for i := 0; i < len(n.FilesOrder); i += 1 {
		k := n.FilesOrder[i]
		files[i] = filesWithIds[k]
	}

	n.Files = files
}
