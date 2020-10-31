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

	Node     *repository.NodeRepository
	User     *repository.UserRepository
	File     *repository.FileRepository
	Meta     *repository.MetaRepository
	Social   *repository.SocialRepository
	NodeView *repository.NodeViewRepository
	Message  *repository.MessageRepository
	Tag      *repository.TagRepository
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
	)

	return &DB{
		DB:       db,
		Node:     new(repository.NodeRepository).Init(db),
		User:     new(repository.UserRepository).Init(db),
		File:     new(repository.FileRepository).Init(db),
		Meta:     new(repository.MetaRepository).Init(db),
		Social:   new(repository.SocialRepository).Init(db),
		NodeView: new(repository.NodeViewRepository).Init(db),
		Message:  new(repository.MessageRepository).Init(db),
		Tag:      new(repository.TagRepository).Init(db),
	}, nil
}
