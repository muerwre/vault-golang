package repository

import "github.com/jinzhu/gorm"

type SearchRepository struct {
	db gorm.DB
}

func (sr *SearchRepository) Init(db *gorm.DB) *SearchRepository {
	sr.db = *db
	return sr
}
