package controller

import (
	"context"
	"github.com/muerwre/vault-golang/internal/db"
	"github.com/muerwre/vault-golang/internal/service/notification/dto"
	"github.com/muerwre/vault-golang/internal/service/notification/usecase"
	"github.com/muerwre/vault-golang/internal/service/vk/controller"
	"github.com/sirupsen/logrus"
)

type NotificationConsumer interface {
	Consume(notification *dto.NotificationDto) error
	Name() string
}

type NotificationService struct {
	Chan         chan *dto.NotificationDto
	notification usecase.NotificationServiceUsecase
	log          *logrus.Logger
	consumers    []NotificationConsumer
}

func (n *NotificationService) Init(db db.DB, log *logrus.Logger) *NotificationService {
	n.Chan = make(chan *dto.NotificationDto, 255)
	n.notification = *new(usecase.NotificationServiceUsecase).Init(db)
	n.log = log

	n.consumers = []NotificationConsumer{
		NewUserNotificationConsumer(db, log),
		controller.NewVkNotificationConsumer(db, log),
	}
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

			for _, v := range n.consumers {
				if err := v.Consume(item); err != nil {
					logrus.Warnf("Failed to consume at %s: %+v", v.Name(), err)
				}
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
