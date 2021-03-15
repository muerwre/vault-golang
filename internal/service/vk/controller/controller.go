package controller

import (
	"context"
	"github.com/muerwre/vault-golang/internal/db"
	"github.com/sirupsen/logrus"
	"time"
)

type VkNotificationService struct {
	config VkNotificationsConfig
	db     db.DB
	log    *logrus.Logger
}

// New returns new instance of VkNotificationService
func New(config VkNotificationsConfig, db db.DB, log *logrus.Logger) *VkNotificationService {
	return new(VkNotificationService).init(config, db, log)
}

func (s *VkNotificationService) init(config VkNotificationsConfig, db db.DB, log *logrus.Logger) *VkNotificationService {
	s.config = config
	s.db = db
	s.log = log

	return s
}

func (s VkNotificationService) Watch(ctx context.Context) {
	if !s.config.Enabled {
		s.log.Debug("VkNotificationService not started")
		return
	}

	s.log.Info("VkNotificationService started")

	for {
		select {
		case <-ctx.Done():
			s.log.Warn("VkNotificationService stopped")
			return
		case <-time.After(time.Minute * time.Duration(s.config.Delay)):
			continue
		}
	}
}
