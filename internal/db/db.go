package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" // needed for gorm
	"github.com/muerwre/vault-golang/internal/db/models"
	repository2 "github.com/muerwre/vault-golang/internal/db/repository"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type DB struct {
	*gorm.DB

	Node                 *repository2.NodeRepository
	User                 *repository2.UserRepository
	File                 *repository2.FileRepository
	Meta                 *repository2.MetaRepository
	Social               *repository2.OauthRepository
	NodeView             *repository2.NodeViewRepository
	Message              *repository2.MessageRepository
	Tag                  *repository2.TagRepository
	Notification         *repository2.NotificationRepository
	Search               *repository2.SearchRepository
	Comment              *repository2.CommentRepository
	MessageView          *repository2.MessageViewRepository
	NotificationSettings *repository2.NotificationSettingsRepository
	AppNotification      *repository2.AppNotificationRepository
}

func New() (*DB, error) {
	config, err := InitConfig()

	if err != nil {
		return nil, errors.Wrap(err, "Cant read config")
	}

	db, err := gorm.Open("mysql", config.URI)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to connect to database")
	}

	if config.Debug {
		db.LogMode(config.Debug)
	}

	logrus.Info("Connected to db")

	db.AutoMigrate(
		&models.User{},
		&models.File{},
		&models.Node{},
		&models.Tag{},
		&models.NodeView{},
		&models.Comment{},
		&models.Token{},
		&models.Social{},
		&models.Message{},
		&models.MessageView{},
		&models.RestoreCode{},
		&models.Embed{},
		&models.NodeWatch{},
		&models.Notification{},
		&models.NotificationSettings{},
		&models.AppNotification{},
	)

	return &DB{
		DB:                   db,
		Node:                 new(repository2.NodeRepository).Init(db),
		User:                 new(repository2.UserRepository).Init(db),
		File:                 new(repository2.FileRepository).Init(db),
		Meta:                 new(repository2.MetaRepository).Init(db),
		Social:               new(repository2.OauthRepository).Init(db),
		NodeView:             new(repository2.NodeViewRepository).Init(db),
		Message:              new(repository2.MessageRepository).Init(db),
		Tag:                  new(repository2.TagRepository).Init(db),
		Search:               new(repository2.SearchRepository).Init(db),
		Comment:              new(repository2.CommentRepository).Init(db),
		MessageView:          new(repository2.MessageViewRepository).Init(db),
		Notification:         new(repository2.NotificationRepository).Init(db),
		NotificationSettings: new(repository2.NotificationSettingsRepository).Init(db),
		AppNotification:      new(repository2.AppNotificationRepository).Init(db),
	}, nil
}
