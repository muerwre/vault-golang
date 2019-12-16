package models

import "github.com/jinzhu/gorm"

type FileMetadata struct {
}

type File struct {
	*gorm.Model

	Name     string
	OrigName string `json:"-"`
	Path     string
	FullPath string
	Url      string
	Size     int
	Type     string
	Target   string
	// User     User `gorm:"foreignkey:UserID"`
	UserID uint `gorm:"column:userId"`
	// Metadata FileMetadata
}
