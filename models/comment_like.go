package models

type CommentLike struct {
	*Model

	Text       string      `json:"text"`
	FilesOrder StringArray `gorm:"column:files_order;type:longtext;" json:"files_order"`
	Files      []*File     `gorm:"many2many:comment_files;" json:"files"`
}
