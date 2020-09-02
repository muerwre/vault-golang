package controllers

import (
	"github.com/muerwre/vault-golang/controllers/usecase"
	"github.com/muerwre/vault-golang/db"
)

type NotificationController struct {
	db      db.DB
	usecase usecase.NotificationUsecase
}

func (c *NotificationController) Init(db db.DB) *NotificationController {
	c.db = db
	c.usecase = *new(usecase.NotificationUsecase).Init(db)

	return c
}
