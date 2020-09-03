package notify

import (
	"github.com/muerwre/vault-golang/db"
	"github.com/sirupsen/logrus"
)

type Notifier struct {
	db   db.DB
	Chan chan *NotifierItem
}

func (n *Notifier) Init(db db.DB) *Notifier {
	n.db = db
	n.Chan = make(chan *NotifierItem, 255)

	return n
}

func (n *Notifier) Listen() {
	logrus.Info("Notifier routine started")

	for {
		select {
		case item, ok := <-n.Chan:
			if !ok {
				logrus.Warnf("Notifier channel closed")
				return
			}

			switch item.Type {
			case NotifierTypeNodeCreate:
				n.OnNodeCreate(*item)
			case NotifierTypeNodeDelete:
				n.OnNodeDelete(*item)
			case NotifierTypeNodeRestore:
				n.OnNodeCreate(*item)
			default:
				logrus.WithField("item", item).Warnf("Got unknown notification of type %s", item.Type)
			}
		}
	}
}

func (n *Notifier) Done() {
	close(n.Chan)
}
