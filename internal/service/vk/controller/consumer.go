package controller

import (
	"github.com/muerwre/vault-golang/internal/db"
	"github.com/muerwre/vault-golang/internal/db/models"
	"github.com/muerwre/vault-golang/internal/service/notification/constants"
	"github.com/muerwre/vault-golang/internal/service/notification/dto"
	"github.com/sirupsen/logrus"
)

const AppId string = "vk"

type VkNotificationConsumer struct {
	db  db.DB
	log *logrus.Logger
}

func NewVkNotificationConsumer(db db.DB, log *logrus.Logger) *VkNotificationConsumer {
	return new(VkNotificationConsumer).Init(db, log)
}

func (c *VkNotificationConsumer) Init(db db.DB, log *logrus.Logger) *VkNotificationConsumer {
	c.db = db
	c.log = log
	return c
}

func (c VkNotificationConsumer) Name() string {
	return "VkNotificationConsumer"
}

func (c VkNotificationConsumer) Consume(item *dto.NotificationDto) error {
	switch item.Type {
	case constants.NotifierTypeNodeCreate, constants.NotifierTypeNodeRestore:
		if err := c.db.AppNotification.Create(AppId, item.ItemId, models.NotificationTypeNode); err != nil {
			return err
		}
		break
	case constants.NotifierTypeNodeDelete:
		if err := c.db.AppNotification.FindAndDelete(AppId, item.ItemId, models.NotificationTypeNode); err != nil {
			return err
		}
		break
	}

	c.log.Infof("%s received notification %+v", c.Name(), item)
	return nil
}
