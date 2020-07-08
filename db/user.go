package db

import (
	"errors"
	"time"

	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
)

func (d *DB) GetUserByToken(t string) (user *models.User, err error) {
	token := &models.Token{}

	d.Preload("User").Preload("User.Photo").Preload("User.Cover").First(&token, "token = ?", t)

	if token.ID == 0 || token.User.ID == 0 {
		return nil, errors.New(codes.UserNotFound)
	}

	return token.User, nil
}

func (d *DB) GetUserByUsername(n string) (user *models.User, err error) {
	user = &models.User{}

	d.Preload("Photo").Preload("Cover").First(&user, "username = ?", n)

	if user.ID == 0 {
		return nil, errors.New(codes.UserNotFound)
	}

	return user, nil
}

func (d *DB) GenerateTokenFor(u *models.User) *models.Token {
	token := &models.Token{
		UserID: u.ID,
	}
	token.New(u.Username)

	d.Create(&token)

	return token
}

func (d *DB) GetUserNewMessages(user models.User, exclude int, last string) (messages []models.Message, err error) {
	foundation, _ := time.Parse(time.RFC3339, "2019-11-10T10:10:22.717Z")

	sq := d.Select("*").
		Table("message").
		Select("MAX(message.id)").
		Joins("LEFT JOIN message_view view ON view.dialogId = message.fromId AND view.userId = ?", user.ID).
		Where("message.toId = ? AND message.fromId != ?", user.ID, user.ID).
		Where("(view.viewed < message.created_at OR view.viewed IS NULL)").
		Where("message.created_at > ?", foundation).
		Where("message.deleted_at IS NULL").
		Group("message.fromId")

	if exclude > 0 {
		sq = sq.Where("message.fromId != ?", exclude)
	}

	if since, err := time.Parse(time.RFC3339, last); err != nil {
		sq = sq.Where("message.created_at > ", since)
	}

	d.Select("*").
		Table("message").
		Where("message.id IN (?)", sq.SubQuery()).
		Order("message.created_at").
		Limit(10).
		Preload("From").
		Preload("To").
		Find(&messages)

	return messages, nil
}
