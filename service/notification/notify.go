package notification

import (
	"github.com/muerwre/vault-golang/db"
	"github.com/sirupsen/logrus"
)

type NotificationService struct {
	db   db.DB
	Chan chan *NotifierItem
}

func (n *NotificationService) Init(db db.DB) *NotificationService {
	n.db = db
	n.Chan = make(chan *NotifierItem, 255)

	return n
}

func (n *NotificationService) Listen() {
	logrus.Info("NotificationService routine started")

	for {
		select {
		case item, ok := <-n.Chan:
			if !ok {
				logrus.Warnf("NotificationService channel closed")
				return
			}

			switch item.Type {
			case NotifierTypeNodeCreate, NotifierTypeNodeRestore:
				n.OnNodeCreate(*item)
			case NotifierTypeNodeDelete:
				n.OnNodeDelete(*item)
			case NotifierTypeCommentCreate, NotifierTypeCommentRestore:
				n.OnCommentCreate(*item)
			case NotifierTypeCommentDelete:
				n.OnCommentDelete(*item)
			default:
				logrus.WithField("item", item).Warnf("Got unknown notification of type %s", item.Type)
			}
		}
	}
}

func (n *NotificationService) Done() {
	close(n.Chan)
}
