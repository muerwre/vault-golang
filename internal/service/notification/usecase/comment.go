package usecase

import (
	"github.com/muerwre/vault-golang/internal/db/models"
	"github.com/muerwre/vault-golang/internal/service/notification/dto"
	"github.com/sirupsen/logrus"
)

func (n NotificationServiceUsecase) CreateUserNotificationsOnCommentCreate(item dto.NotificationDto) {
	c, err := n.node.GetCommentByIdWithDeleted(item.ItemId)

	if err != nil {
		logrus.Warnf("Comment with id %s not found", item.ItemId)
		return
	}

	recipients, err := n.node.GetNodeWatchers(*c.NodeID)

	if err != nil {
		logrus.Warnf("Can't get watchers for node %d", c.NodeID)
		return
	}

	for _, v := range recipients {
		notification := &models.Notification{
			Type:   models.NotificationTypeComment,
			ItemID: item.ItemId,
			UserID: v,
			Time:   item.CreatedAt,
		}

		if err := n.notification.Create(notification); err != nil {
			logrus.Warnf("Can't perform CreateUserNotificationsOnCommentCreate: %s", err.Error())
		}
	}
}

func (n NotificationServiceUsecase) ClearUserNotificationsOnCommentDelete(item dto.NotificationDto) {
	if err := n.notification.DeleteByTypeAndId(models.NotificationTypeComment, item.ItemId); err != nil {
		logrus.Warnf("Can't perform ClearUserNotificationsOnCommentDelete notification: %s", err.Error())
	}
}
