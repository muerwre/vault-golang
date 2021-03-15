package usecase

import (
	"github.com/muerwre/vault-golang/internal/db/models"
	"github.com/muerwre/vault-golang/internal/service/notification/dto"
	"github.com/sirupsen/logrus"
)

func (n NotificationServiceUsecase) CreateUserNotificationsOnNodeCreate(item dto.NotificationDto) error {
	recipients, err := n.notificationSettings.GetFlowWatchers()

	if err != nil {
		logrus.Warnf("Can't get watchers for node %d", item.ItemId)
		return err
	}

	for _, v := range recipients {
		notification := &models.Notification{
			Type:   models.NotificationTypeNode,
			ItemID: item.ItemId,
			UserID: v,
			Time:   item.CreatedAt,
		}

		if err := n.notification.Create(notification); err != nil {
			logrus.Warnf("Can't perform CreateUserNotificationsOnNodeCreate: %s", err.Error())
		}
	}

	return nil
}

func (n NotificationServiceUsecase) ClearUserNotificationsOnNodeDelete(item dto.NotificationDto) error {
	if err := n.notification.DeleteByTypeAndId(models.NotificationTypeNode, item.ItemId); err != nil {
		logrus.Warnf("Can't perform ClearUserNotificationsOnNodeDelete: %s", err.Error())
		return err
	}

	return nil
}
