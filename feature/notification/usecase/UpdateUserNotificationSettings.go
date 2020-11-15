package usecase

import (
	"github.com/muerwre/vault-golang/db/models"
	"github.com/muerwre/vault-golang/feature/notification/request"
)

func (u NotificationUsecase) UpdateUserNotificationSettings(uid uint, req *request.NotificationSettingsRequest) (*models.NotificationSettings, error) {
	return u.notificationSettings.UpdateSettings(uid, req.ToModel())
}
