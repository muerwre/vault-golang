package notify

import (
	"github.com/muerwre/vault-golang/models"
	"github.com/sirupsen/logrus"
)

func (n Notifier) OnNodeCreate(item NotifierItem) {
	// TODO: fetch recipients here
	recipients, err := n.db.UserRepository.GetFlowWatchers()

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
	if err := n.db.NotificationRepository.DeleteByTypeAndId(item.Type, item.ItemId); err != nil {
		logrus.Warnf("Can't perform OnNodeDelete: %s", err.Error())
	}
}

func (n Notifier) OnNodeRestore(item NotifierItem) {
	if err := n.db.NotificationRepository.RestoreByTypeAndId(models.NotificationTypeNode, item.ItemId); err != nil {
		logrus.Warnf("Can't perform OnNodeRestore: %s", err.Error())
	}
}
