package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/db/models"
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

func (r *MessageRepository) LoadUnscopedMessageWithUsers(ID uint) (*models.Message, error) {
	message := &models.Message{}

	q := r.db.Where("id = ?", ID).
		Unscoped().
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

func (r *MessageRepository) Delete(ID uint) error {
	return r.db.Delete(&models.Message{}, "id = ?", ID).Error
}

func (r *MessageRepository) Restore(ID uint) error {
	return r.db.Model(&models.Message{}).Unscoped().Where("id = ?", ID).Update("deleted_at", nil).Error
}
