package notify

import (
	"fmt"
	"github.com/muerwre/vault-golang/db"
	"github.com/sirupsen/logrus"
	"time"
)

type AnyNotification interface {
	GetType() string
	GetContent() interface{}
	GetCreatedAt() time.Time
}

type Notifier struct {
	db   db.DB
	Chan chan *AnyNotification
}

func (n *Notifier) Init(db db.DB) *Notifier {
	n.db = db
	n.Chan = make(chan *AnyNotification, 255)

	return n
}

func (n *Notifier) Listen() {
	logrus.Info("Notifier routine started")

	for {
		select {
		case m, ok := <-n.Chan:
			if !ok {
				logrus.Warnf("Notifier channel closed")
				return
			}

			fmt.Printf("Got notification: %+v", m)
		}
	}
}

func (n *Notifier) Done() {
	close(n.Chan)
}
