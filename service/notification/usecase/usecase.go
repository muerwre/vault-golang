package usecase

import (
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/db/repository"
)

type NotificationServiceUsecase struct {
	user         repository.UserRepository
	notification repository.NotificationRepository
	node         repository.NodeRepository
}

func (n *NotificationServiceUsecase) Init(db db.DB) *NotificationServiceUsecase {
	n.user = *db.User
	n.notification = *db.NotificationRepository
	n.node = *db.Node
	return n
}
