package repository

import "github.com/jinzhu/gorm"

type AppNotificationRepository struct {
	db *gorm.DB
}

func (cr *AppNotificationRepository) Init(db *gorm.DB) *AppNotificationRepository {
	cr.db = db
	return cr
}
