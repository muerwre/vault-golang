package usecase

import "github.com/muerwre/vault-golang/db"

type NotificationUsecase struct {
	db db.DB
}

func (u *NotificationUsecase) Init(db db.DB) *NotificationUsecase {
	u.db = db
	return u
}
