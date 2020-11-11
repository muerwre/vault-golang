package usecase

import (
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/db/repository"
)

type NotificationUsecase struct {
	notification repository.NotificationRepository
}

func (u *NotificationUsecase) Init(db db.DB) *NotificationUsecase {
	u.notification = *db.NotificationRepository
	return u
}
