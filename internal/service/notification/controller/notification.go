package controller

import (
	"context"
	"github.com/muerwre/vault-golang/internal/db"
	"github.com/muerwre/vault-golang/internal/service/notification/constants"
	"github.com/muerwre/vault-golang/internal/service/notification/dto"
	"github.com/muerwre/vault-golang/internal/service/notification/usecase"
	"github.com/sirupsen/logrus"
)

type NotificationService struct {
	Chan         chan *dto.NotificationDto
	notification usecase.NotificationServiceUsecase
	log          *logrus.Logger
}

func (n *NotificationService) Init(db db.DB, log *logrus.Logger) *NotificationService {
	n.Chan = make(chan *dto.NotificationDto, 255)
	n.notification = *new(usecase.NotificationServiceUsecase).Init(db)
	n.log = log

	return n
}

func (n *NotificationService) Listen(ctx context.Context) {
	logrus.Info("NotificationService started")

	for {
		select {
		case <-ctx.Done():
			close(n.Chan)
			n.log.Info("NotificationService stopped")
			return
		case item, ok := <-n.Chan:
			if !ok {
				logrus.Warnf("NotificationService channel closed")
				return
			}

			switch item.Type {
			case constants.NotifierTypeNodeCreate, constants.NotifierTypeNodeRestore:
				n.notification.CreateUserNotificationsOnNodeCreate(*item)
			case constants.NotifierTypeNodeDelete:
				n.notification.ClearUserNotificationsOnNodeDelete(*item)
			case constants.NotifierTypeCommentCreate, constants.NotifierTypeCommentRestore:
				n.notification.CreateUserNotificationsOnCommentCreate(*item)
			case constants.NotifierTypeCommentDelete:
				n.notification.ClearUserNotificationsOnCommentDelete(*item)
			default:
				logrus.WithField("item", item).Warnf("Got unknown notification of type %s", item.Type)
			}
		}
	}
}

func (n *NotificationService) Receive(notification *dto.NotificationDto) error {
	n.Chan <- notification
	return nil
}

func (n *NotificationService) Done() {
	close(n.Chan)
}
