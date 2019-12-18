package db

import (
	"errors"

	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
)

func (d *DB) GetUserByToken(t string) (user *models.User, err error) {
	token := &models.Token{}

	d.Preload("User").Preload("User.Photo").Preload("User.Cover").First(&token, "token = ?", t)

	if token.ID == 0 || token.User.ID == 0 {
		return nil, errors.New(codes.USER_NOT_FOUND)
	}

	return token.User, nil
}

func (d *DB) GetUserByUsername(n string) (user *models.User, err error) {
	user = &models.User{}

	d.Preload("Photo").Preload("Cover").First(&user, "username = ?", n)

	if user.ID == 0 {
		return nil, errors.New(codes.USER_NOT_FOUND)
	}

	return user, nil
}
