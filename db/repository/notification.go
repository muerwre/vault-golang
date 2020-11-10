package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/db/models"
)

type NotificationRepository struct {
	db *gorm.DB
}

func (r *NotificationRepository) Init(db *gorm.DB) *NotificationRepository {
	r.db = db
	return r
}

func (r NotificationRepository) Create(notification *models.Notification) error {
	return nil // TODO: Just delete this when you want notifications
	return r.db.Model(&notification).Create(&notification).Error
}

func (r NotificationRepository) DeleteByTypeAndId(t string, id uint) error {
	return nil // TODO: Just delete this when you want notifications
	item := &models.Notification{}
	if err := r.db.Unscoped().Model(&item).Where("type = ? AND itemId = ?", t, id).First(&item).Error; err != nil {
		return err
	}
	return r.db.Unscoped().Delete(&item).Error
}