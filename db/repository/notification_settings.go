package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/db/models"
)

type NotificationSettingsRepository struct {
	db *gorm.DB
}

func (nr *NotificationSettingsRepository) Init(db *gorm.DB) *NotificationSettingsRepository {
	nr.db = db
	return nr
}

func (nr NotificationSettingsRepository) GetFlowWatchers() ([]uint, error) {
	var users []*models.NotificationSettings

	if err := nr.db.Model(&models.NotificationSettings{}).Where("subscribed_to_flow = ?", true).Find(&users).Error; err != nil {
		return nil, err
	}

	ids := make([]uint, len(users))

	for k, v := range users {
		ids[k] = *v.UserID
	}

	return ids, nil
}
