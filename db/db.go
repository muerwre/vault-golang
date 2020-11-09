package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" // needed for gorm
	messageRepository "github.com/muerwre/vault-golang/feature/message/repository"
	metaRepository "github.com/muerwre/vault-golang/feature/meta/repository"
	nodeRepository "github.com/muerwre/vault-golang/feature/node/repository"
	notificationRepository "github.com/muerwre/vault-golang/feature/notification/repository"
	oauthRepository "github.com/muerwre/vault-golang/feature/oauth/repository"
	"github.com/muerwre/vault-golang/feature/search/repository"
	tagRepository "github.com/muerwre/vault-golang/feature/tag/repository"
	fileRepository "github.com/muerwre/vault-golang/feature/upload/repository"
	userRepository "github.com/muerwre/vault-golang/feature/user/repository"
	"github.com/muerwre/vault-golang/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type DB struct {
	*gorm.DB

	Node                   *nodeRepository.NodeRepository
	User                   *userRepository.UserRepository
	File                   *fileRepository.FileRepository
	Meta                   *metaRepository.MetaRepository
	Social                 *oauthRepository.OauthRepository
	NodeView               *nodeRepository.NodeViewRepository
	Message                *messageRepository.MessageRepository
	Tag                    *tagRepository.TagRepository
	NotificationRepository *notificationRepository.NotificationRepository
	Search                 *repository.SearchRepository
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
	)

	return &DB{
		DB:                     db,
		Node:                   new(nodeRepository.NodeRepository).Init(db),
		User:                   new(userRepository.UserRepository).Init(db),
		File:                   new(fileRepository.FileRepository).Init(db),
		Meta:                   new(metaRepository.MetaRepository).Init(db),
		Social:                 new(oauthRepository.OauthRepository).Init(db),
		NodeView:               new(nodeRepository.NodeViewRepository).Init(db),
		Message:                new(messageRepository.MessageRepository).Init(db),
		Tag:                    new(tagRepository.TagRepository).Init(db),
		NotificationRepository: new(notificationRepository.NotificationRepository).Init(db),
		Search:                 new(repository.SearchRepository).Init(db),
	}, nil
}
