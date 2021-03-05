package usecase

import (
	"github.com/muerwre/vault-golang/internal/db"
	"github.com/muerwre/vault-golang/internal/db/repository"
)

type NotificationUsecase struct {
	notification         repository.NotificationRepository
	notificationSettings repository.NotificationSettingsRepository
}

func (u *NotificationUsecase) Init(db db.DB) *NotificationUsecase {
	u.notification = *db.Notification
	u.notificationSettings = *db.NotificationSettings
	return u
}
