package controller

import (
	"github.com/muerwre/vault-golang/internal/db"
	"github.com/muerwre/vault-golang/internal/service/notification/constants"
	"github.com/muerwre/vault-golang/internal/service/notification/dto"
	"github.com/muerwre/vault-golang/internal/service/notification/usecase"
	"github.com/sirupsen/logrus"
)

type UserNotificationConsumer struct {
	notification usecase.NotificationServiceUsecase
	log          *logrus.Logger
}

func NewUserNotificationConsumer(db db.DB, log *logrus.Logger) *UserNotificationConsumer {
	return new(UserNotificationConsumer).Init(db, log)
}

func (s *UserNotificationConsumer) Init(db db.DB, log *logrus.Logger) *UserNotificationConsumer {
	s.notification = *new(usecase.NotificationServiceUsecase).Init(db)
	s.log = log
	return s
}

func (s UserNotificationConsumer) Name() string {
	return "UserNotificationConsumer"
}

func (s UserNotificationConsumer) Consume(item *dto.NotificationDto) error {
	switch item.Type {
	case constants.NotifierTypeNodeCreate, constants.NotifierTypeNodeRestore:
		if err := s.notification.CreateUserNotificationsOnNodeCreate(*item); err != nil {
			return err
		}
	case constants.NotifierTypeNodeDelete:
		if err := s.notification.ClearUserNotificationsOnNodeDelete(*item); err != nil {
			return err
		}
	case constants.NotifierTypeCommentCreate, constants.NotifierTypeCommentRestore:
		if err := s.notification.CreateUserNotificationsOnCommentCreate(*item); err != nil {
			return err
		}
	case constants.NotifierTypeCommentDelete:
		if err := s.notification.ClearUserNotificationsOnCommentDelete(*item); err != nil {
			return err
		}
	default:
		logrus.WithField("item", item).Warnf("Got unknown notification of type %s", item.Type)
	}

	return nil
}
