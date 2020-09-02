package repository

import (
	"github.com/jinzhu/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func (r *NotificationRepository) Init(db *gorm.DB) *NotificationRepository {
	r.db = db
	return r
}
