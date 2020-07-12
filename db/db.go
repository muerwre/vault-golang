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

	NodeRepository *repository.NodeRepository
	UserRepository *repository.UserRepository
	FileRepository *repository.FileRepository
	MetaRepository *repository.MetaRepository
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

	nr := &repository.NodeRepository{}
	ur := &repository.UserRepository{}
	fr := &repository.FileRepository{}
	mr := &repository.MetaRepository{}

	nr.Init(db)
	ur.Init(db)
	fr.Init(db)
	mr.Init(db)

	return &DB{
		DB:             db,
		NodeRepository: nr,
		UserRepository: ur,
		FileRepository: fr,
		MetaRepository: mr,
	}, nil
}
