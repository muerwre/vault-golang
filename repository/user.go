package repository

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
	"time"
)

type UserRepository struct {
	db *gorm.DB
}

func (ur *UserRepository) Init(db *gorm.DB) *UserRepository {
	ur.db = db
	return ur
}

func (ur UserRepository) Create(user *models.User) (err error) {
	ur.db.Create(&user)
	return nil
}

func (ur UserRepository) Save(user *models.User) (err error) {
	ur.db.Save(&user)
	return nil
}

func (ur UserRepository) GetByToken(t string) (user *models.User, err error) {
	token := &models.Token{}

	ur.db.Preload("User").Preload("User.Photo").Preload("User.Cover").First(&token, "token = ?", t)

	if token.ID == 0 || token.User.ID == 0 {
		return nil, errors.New(codes.UserNotFound)
	}

	return token.User, nil
}

func (ur UserRepository) GetByUsername(n string) (user *models.User, err error) {
	user = &models.User{}

	ur.db.Preload("Photo").Preload("Cover").First(&user, "username = ?", n)

	if user.ID == 0 {
		return nil, errors.New(codes.UserNotFound)
	}

	return user, nil
}

func (ur UserRepository) GetByEmail(n string) (user *models.User, err error) {
	user = &models.User{}

	ur.db.Preload("Photo").Preload("Cover").First(&user, "email = ?", n)

	if user.ID == 0 {
		return nil, errors.New(codes.UserNotFound)
	}

	return user, nil
}

func (ur UserRepository) GenerateTokenFor(u *models.User) *models.Token {
	token := &models.Token{UserID: &u.ID}
	token.New(u.Username)

	ur.db.Create(&token)

	return token
}

func (ur UserRepository) GetUserNewMessages(user models.User, exclude int, last string) (messages []models.Message, err error) {
	foundation, _ := time.Parse(time.RFC3339, "2019-11-10T10:10:22.717Z")

	sq := ur.db.Select("*").
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

	ur.db.Select("*").
		Table("message").
		Where("message.id IN (?)", sq.SubQuery()).
		Order("message.created_at").
		Limit(10).
		Preload("From").
		Preload("To").
		Find(&messages)

	return messages, nil
}

func (ur UserRepository) GetTotalCount() (count int) {
	ur.db.Model(&models.User{}).Count(&count)
	return
}

func (ur UserRepository) GetAliveCount() (count int) {
	ur.db.Model(&models.User{}).Where("user.last_seen > NOW() - INTERVAL 40 DAY").Count(&count)
	return
}

func (ur UserRepository) UpdateLastSeen(user *models.User) {
	user.LastSeen = time.Now()
	ur.db.Save(&user)
}
