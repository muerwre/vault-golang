package models

type CommentLike struct {
	*Model

	Text       string           `json:"text"`
	FilesOrder CommaStringArray `gorm:"column:files_order;type:longtext;" json:"files_order"`
}
