package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/models"
)

type MessageRepository struct {
	db *gorm.DB
}

func (r *MessageRepository) Init(d *gorm.DB) *MessageRepository {
	r.db = d
	return r
}

func (r *MessageRepository) LoadMessageWithUsers(ID uint) (*models.Message, error) {
	message := &models.Message{}

	q := r.db.Where("id = ?", ID).
		Preload("From").
		Preload("To").
		First(&message)

	return message, q.Error
}

func (r *MessageRepository) CreateMessage(message *models.Message) error {
	return r.db.Model(&models.Message{}).Create(&message).Error
}

func (r *MessageRepository) SaveMessage(message *models.Message) error {
	return r.db.Model(&models.Message{}).Save(&message).Error
}
