package models

import (
	"time"
)

type Comment struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `sql:"index" json:"-"`

	Text       string         `json:"text"`
	FilesOrder CommaUintArray `gorm:"column:files_order;type:longtext;" json:"files_order"`

	User   *User `json:"user" gorm:"foreignkey:UserID"`
	UserID uint  `gorm:"column:userId" json:"-"`

	Node   *Node   `json:"node" gorm:"foreignkey:NodeID"`
	NodeID uint    `gorm:"column:nodeId" json:"-"`
	Files  []*File `gorm:"many2many:comment_files_file;jointable_foreignkey:commentId;association_jointable_foreignkey:fileId" json:"files"`
}

var COMMENT_FILE_TYPES = struct {
	IMAGE string
	AUDIO string
}{
	IMAGE: "image",
	AUDIO: "audio",
}

// var CommentFiles =

func (Comment) TableName() string {
	return "comment"
}

// SortFiles - sorts comment files according to files_order
func (c *Comment) SortFiles() {
	if len(c.FilesOrder) == 0 || len(c.Files) == 0 {
		return
	}

	filesWithIds := make(map[uint]*File, len(c.FilesOrder))
	files := make([]*File, len(c.FilesOrder))

	for i := 0; i < len(c.Files); i += 1 {
		k := c.Files[i]
		filesWithIds[k.ID] = k
	}

	for i := 0; i < len(c.FilesOrder); i += 1 {
		k := c.FilesOrder[i]
		files[i] = filesWithIds[k]
	}

	c.Files = files
}

// CanBeEditedBy checks if comment can be edited by user
func (c *Comment) CanBeEditedBy(user *User) bool {
	return user.Role == USER_ROLES.ADMIN || c.UserID == user.ID
}
