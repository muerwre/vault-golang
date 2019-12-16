package db

import (
	"errors"

	"github.com/muerwre/vault-golang/models"
)

func (d *DB) GetUserByToken(t string) (user *models.User, err error) {
	token := &models.Token{}

	d.Preload("User").Preload("User.Photo").Preload("User.Cover").First(&token, "token = ?", t)

	if token == nil || token.User == nil {
		return nil, errors.New("Not found")
	}

	return token.User, nil
}
