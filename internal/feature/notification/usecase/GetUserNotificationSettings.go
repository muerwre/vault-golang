package usecase

import (
	"github.com/muerwre/vault-golang/internal/db/models"
)

func (u NotificationUsecase) GetUserNotificationSettings(uid uint) (*models.NotificationSettings, error) {
	return u.notificationSettings.GetForUserId(uid)
}
