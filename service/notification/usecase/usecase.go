package usecase

import (
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/db/repository"
)

type NotificationServiceUsecase struct {
	user                 repository.UserRepository
	notification         repository.NotificationRepository
	notificationSettings repository.NotificationSettingsRepository
	node                 repository.NodeRepository
}

func (n *NotificationServiceUsecase) Init(db db.DB) *NotificationServiceUsecase {
	n.user = *db.User
	n.notification = *db.Notification
	n.node = *db.Node
	n.notificationSettings = *db.NotificationSettings
	return n
}
