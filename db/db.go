package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" // needed for gorm
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type DB struct {
	*gorm.DB
}

func New() (*DB, error) {
	config, err := InitConfig()

	if err != nil {
		return nil, errors.Wrap(err, "Cant read config")
	}

	db, err := gorm.Open("mysql", config.URI)

	if config.Debug {
		db.LogMode(config.Debug)
	}

	if err != nil {
		return nil, errors.Wrap(err, "Unable to connect to database")
	}

	logrus.Info("Connected to db")

	// db.AutoMigrate(&model.User{}, &model.Route{})

	return &DB{db}, nil
}
