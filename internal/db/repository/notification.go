package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/internal/db/models"
)

type NotificationRepository struct {
	db *gorm.DB
}

func (r *NotificationRepository) Init(db *gorm.DB) *NotificationRepository {
	r.db = db
	return r
}

func (r NotificationRepository) Create(notification *models.Notification) error {
	return r.db.Model(&notification).Create(&notification).Error
}

func (r NotificationRepository) DeleteByTypeAndId(t string, id uint) error {
	return r.db.Unscoped().Delete(&models.Notification{}, "type = ? AND itemId = ?", t, id).Error
}
