package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/internal/db/models"
	"time"
)

type AppNotificationRepository struct {
	db *gorm.DB
}

func (cr *AppNotificationRepository) Init(db *gorm.DB) *AppNotificationRepository {
	cr.db = db
	return cr
}

func (cr AppNotificationRepository) FindByTypeAndId(app string, id uint, t string) (*models.AppNotification, error) {
	res := &models.AppNotification{}

	if err := cr.db.First(
		res,
		"app = ? AND item_id = ? AND type = ?",
		app,
		id,
		t,
	).Error; err != nil {
		return nil, err
	}

	return res, nil
}
func (cr AppNotificationRepository) Create(app string, id uint, t string) error {
	// Skip it as already exist
	if _, err := cr.FindByTypeAndId(app, id, t); err == nil {
		return nil
	}

	item := &models.AppNotification{
		App:    app,
		ItemID: id,
		Type:   t,
	}

	return cr.db.Create(item).Error
}

func (cr AppNotificationRepository) FindAndDeleteUnsent(app string, id uint, t string) error {
	return cr.db.Delete(
		&models.AppNotification{},
		"app = ? AND item_id = ? AND type = ? AND sent_at IS NULL",
		app,
		id,
		t,
	).Error
}

func (cr AppNotificationRepository) FindLatest(laterThan time.Time, earlierThan time.Time) ([]models.AppNotification, error) {
	res := &[]models.AppNotification{}

	err := cr.db.Find(
		res,
		"created_at > ? AND created_at < ?",
		laterThan,
		earlierThan,
	).Error

	return *res, err
}
