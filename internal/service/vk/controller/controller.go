package controller

import (
	"context"
	"fmt"
	"github.com/muerwre/vault-golang/internal/db"
	"github.com/muerwre/vault-golang/internal/db/models"
	"github.com/muerwre/vault-golang/internal/service/notification/dto"
	"github.com/muerwre/vault-golang/pkg/vk"
	"github.com/sirupsen/logrus"
	"time"
)

type VkNotificationService struct {
	config VkNotificationsConfig

	db  db.DB
	log *logrus.Logger
	vk  *vk.Vk
}

// New returns new instance of VkNotificationService
func New(config VkNotificationsConfig, db db.DB, log *logrus.Logger) *VkNotificationService {
	return new(VkNotificationService).init(config, db, log)
}

func (s *VkNotificationService) init(config VkNotificationsConfig, db db.DB, log *logrus.Logger) *VkNotificationService {
	s.config = config
	s.db = db
	s.log = log
	s.vk = vk.NewVk(config.ApiKey, config.GroupId, log)

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
			s.log.Info("VkNotificationService stopped")
			return
		case <-time.After(time.Minute * time.Duration(s.config.CooldownMins)):
			if err := s.publishUnsentNotifications(ctx); err != nil {
				s.log.Warnf("Failed to publish notifications: %+v", err)
			}

			continue
		}
	}
}

func (s VkNotificationService) publishUnsentNotifications(ctx context.Context) error {
	laterThan := time.Now().Add(time.Duration(-1*int32(s.config.CooldownMins)) * time.Minute)
	earlierThan := time.Now().Add(time.Duration(-1*int32(s.config.PurgeAfterDays)) * time.Hour * 24)

	latest, err := s.db.AppNotification.FindLatest(AppId, laterThan, earlierThan)

	if err != nil {
		return err
	}

	for _, v := range latest {
		switch v.Type {
		case models.NotificationTypeNode:
			if err := s.publishNode(ctx, v.ToDto()); err != nil {
				s.log.Warnf("Can't publish node %d: %+v", v.ItemID, err)
			}
		}
	}

	return nil
}

func (s VkNotificationService) publishNode(ctx context.Context, dto dto.NotificationDto) error {
	node, err := s.db.Node.GetNodeWithUserAndTags(dto.ItemId)
	if err != nil {
		return err
	}

	if !node.IsFlowType() {
		return fmt.Errorf("node %d is not of flow type", node.ID)
	}

	msg, u, thumb := s.getNodeContent(*node)

	if err := s.vk.CreatePost(ctx, msg, u, thumb); err != nil {
		return err
	}

	now := time.Now()
	if err := s.db.AppNotification.SetSent(AppId, dto.ItemId, models.NotificationTypeNode, &now); err != nil {
		s.db.AppNotification.FindAndDeleteUnsent(AppId, dto.ItemId, models.NotificationTypeNode)
	}

	return nil
}

func (s VkNotificationService) getNodeContent(node models.Node) (string, string, string) {
	u := fmt.Sprintf("%s/post%d", s.config.UrlPrefix, node.ID)
	msg := fmt.Sprintf("%s\n~%s\n%s", node.Title, node.User.Username, u)

	if node.Description != "" {
		msg = msg + "\n\n" + node.Description
	}

	return msg, "", node.GetThumbnailPath(s.config.UploadPath)
}
