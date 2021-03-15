package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/internal/db/models"
)

type AppNotificationRepository struct {
	db *gorm.DB
}

func (cr *AppNotificationRepository) Init(db *gorm.DB) *AppNotificationRepository {
	cr.db = db
	return cr
}

func (cr AppNotificationRepository) Create(app string, id uint, t string) error {
	item := &models.AppNotification{
		App:    app,
		ItemID: id,
		Type:   t,
	}

	return cr.db.Create(item).Error
}

func (cr AppNotificationRepository) FindAndDelete(app string, id uint, t string) error {
	return cr.db.Delete(&models.AppNotification{}, "app = ? AND item_id = ? AND type = ?", app, id, t).Error
}
