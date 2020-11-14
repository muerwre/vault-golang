package controller

import (
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/service/notification/constants"
	"github.com/muerwre/vault-golang/service/notification/dto"
	"github.com/muerwre/vault-golang/service/notification/usecase"
	"github.com/sirupsen/logrus"
)

type NotificationService struct {
	Chan         chan *dto.NotificationDto
	notification usecase.NotificationServiceUsecase
}

func (n *NotificationService) Init(db db.DB) *NotificationService {
	n.Chan = make(chan *dto.NotificationDto, 255)
	n.notification = *new(usecase.NotificationServiceUsecase).Init(db)

	return n
}

func (n *NotificationService) Listen() {
	logrus.Info("Notification NotificationService is listening")

	for {
		select {
		case item, ok := <-n.Chan:
			if !ok {
				logrus.Warnf("NotificationService channel closed")
				return
			}

			switch item.Type {
			case constants.NotifierTypeNodeCreate, constants.NotifierTypeNodeRestore:
				n.notification.OnNodeCreate(*item)
			case constants.NotifierTypeNodeDelete:
				n.notification.OnNodeDelete(*item)
			case constants.NotifierTypeCommentCreate, constants.NotifierTypeCommentRestore:
				n.notification.OnCommentCreate(*item)
			case constants.NotifierTypeCommentDelete:
				n.notification.OnCommentDelete(*item)
			default:
				logrus.WithField("item", item).Warnf("Got unknown notification of type %s", item.Type)
			}
		}
	}
}

func (n *NotificationService) Receive(dto *dto.NotificationDto) error {
	n.Chan <- dto
	return nil
}

func (n *NotificationService) Done() {
	close(n.Chan)
}
