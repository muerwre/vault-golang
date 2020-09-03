package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/models"
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
	item := &models.Notification{}

	if err := r.db.Model(&item).First(&item).Error; err != nil {
		return err
	}

	return r.db.Delete(&item).Error
}

func (r NotificationRepository) RestoreByTypeAndId(t string, id uint) error {

	return r.db.Unscoped().Model(&models.Notification{}).
		Where("itemId = ? AND type = ?", id, t).
		Update("deleted_at", nil).Error
}
