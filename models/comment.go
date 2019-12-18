package models

import "fmt"

type Comment struct {
	*CommentLike

	User   *User `json:"user" gorm:"foreignkey:UserID"`
	UserID uint  `gorm:"column:userId" json:"-"`

	Node   *Node   `json:"node" gorm:"foreignkey:NodeID"`
	NodeID uint    `gorm:"column:nodeId" json:"-"`
	Files  []*File `gorm:"many2many:comment_files_file;jointable_foreignkey:commentId;association_jointable_foreignkey:fileId" json:"files"`
}

func (Comment) TableName() string {
	return "comment"
}

func (c *Comment) SortFiles() {
	if len(c.FilesOrder) == 0 || len(c.Files) == 0 {
		return
	}

	filesWithIds := make(map[string]*File, len(c.FilesOrder))
	files := make([]*File, len(c.FilesOrder))

	for i := 0; i < len(c.Files); i += 1 {
		k := c.Files[i]
		filesWithIds[fmt.Sprint(k.ID)] = k
	}

	for i := 0; i < len(c.FilesOrder); i += 1 {
		k := c.FilesOrder[i]
		c.Files[i] = filesWithIds[k]
	}

	c.Files = files
}
