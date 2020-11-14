package usecase

import (
	"github.com/muerwre/vault-golang/db/models"
	"github.com/muerwre/vault-golang/service/notification/dto"
	"github.com/sirupsen/logrus"
)

func (n NotificationServiceUsecase) OnCommentCreate(item dto.NotificationDto) {
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
			logrus.Warnf("Can't perform OnCommentCreate: %s", err.Error())
		}
	}
}

func (n NotificationServiceUsecase) OnCommentDelete(item dto.NotificationDto) {
	if err := n.notification.DeleteByTypeAndId(models.NotificationTypeComment, item.ItemId); err != nil {
		logrus.Warnf("Can't perform OnCommentDelete notification: %s", err.Error())
	}
}
