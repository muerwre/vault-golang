package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" // needed for gorm
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/repository"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type DB struct {
	*gorm.DB

	NodeRepository         *repository.NodeRepository
	UserRepository         *repository.UserRepository
	FileRepository         *repository.FileRepository
	MetaRepository         *repository.MetaRepository
	SocialRepository       *repository.SocialRepository
	NodeViewRepository     *repository.NodeViewRepository
	NotificationRepository *repository.NotificationRepository
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
	)

	return &DB{
		DB:                     db,
		NodeRepository:         new(repository.NodeRepository).Init(db),
		UserRepository:         new(repository.UserRepository).Init(db),
		FileRepository:         new(repository.FileRepository).Init(db),
		MetaRepository:         new(repository.MetaRepository).Init(db),
		SocialRepository:       new(repository.SocialRepository).Init(db),
		NodeViewRepository:     new(repository.NodeViewRepository).Init(db),
		NotificationRepository: new(repository.NotificationRepository).Init(db),
	}, nil
}
