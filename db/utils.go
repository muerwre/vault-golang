package db

import (
	"errors"

	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
)

func (d *DB) GetUserByToken(t string) (user *models.User, err error) {
	token := &models.Token{}

	// .Preload("User.Photo").Preload("User.Cover")
	d.Preload("User").First(&token, "token = ?", t)

	if token == nil || token.User == nil {
		return nil, errors.New(codes.USER_NOT_FOUND)
	}

	return token.User, nil
}

func (d *DB) GetUserByUsername(n string) (user *models.User, err error) {
	user = &models.User{}

	d.Preload("Photo").Preload("Cover").First(&user, "username = ?", n)

	if user == nil {
		return nil, errors.New(codes.USER_NOT_FOUND)
	}

	return user, nil
}
