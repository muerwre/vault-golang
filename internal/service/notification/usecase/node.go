package usecase

import (
	"github.com/muerwre/vault-golang/internal/db/models"
	"github.com/muerwre/vault-golang/internal/service/notification/dto"
	"github.com/sirupsen/logrus"
)

func (n NotificationServiceUsecase) OnNodeCreate(item dto.NotificationDto) {
	recipients, err := n.notificationSettings.GetFlowWatchers()

	if err != nil {
		logrus.Warnf("Can't get watchers for node %d", item.ItemId)
		return
	}

	for _, v := range recipients {
		notification := &models.Notification{
			Type:   models.NotificationTypeNode,
			ItemID: item.ItemId,
			UserID: v,
			Time:   item.CreatedAt,
		}

		if err := n.notification.Create(notification); err != nil {
			logrus.Warnf("Can't perform OnNodeCreate: %s", err.Error())
		}
	}
}

func (n NotificationServiceUsecase) OnNodeDelete(item dto.NotificationDto) {
	if err := n.notification.DeleteByTypeAndId(models.NotificationTypeNode, item.ItemId); err != nil {
		logrus.Warnf("Can't perform OnNodeDelete: %s", err.Error())
	}
}
