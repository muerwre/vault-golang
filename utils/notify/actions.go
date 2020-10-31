package notify

import (
	"github.com/muerwre/vault-golang/models"
	"github.com/sirupsen/logrus"
)

func (n Notifier) OnNodeCreate(item NotifierItem) {
	// TODO: fetch recipients here
	recipients, err := n.db.User.GetFlowWatchers()

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

		if err := n.db.NotificationRepository.Create(notification); err != nil {
			logrus.Warnf("Can't perform OnNodeCreate: %s", err.Error())
		}
	}
}

func (n Notifier) OnNodeDelete(item NotifierItem) {
	if err := n.db.NotificationRepository.DeleteByTypeAndId(models.NotificationTypeNode, item.ItemId); err != nil {
		logrus.Warnf("Can't perform OnNodeDelete: %s", err.Error())
	}
}

func (n Notifier) OnCommentCreate(item NotifierItem) {
	c, err := n.db.Node.GetCommentByIdWithDeleted(item.ItemId)

	if err != nil {
		logrus.Warnf("Comment with id %s not found", item.ItemId)
		return
	}

	recipients, err := n.db.Node.GetNodeWatchers(*c.NodeID)

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

		if err := n.db.NotificationRepository.Create(notification); err != nil {
			logrus.Warnf("Can't perform OnCommentCreate: %s", err.Error())
		}
	}
}

func (n Notifier) OnCommentDelete(item NotifierItem) {
	if err := n.db.NotificationRepository.DeleteByTypeAndId(models.NotificationTypeComment, item.ItemId); err != nil {
		logrus.Warnf("Can't perform OnCommentDelete: %s", err.Error())
	}
}
