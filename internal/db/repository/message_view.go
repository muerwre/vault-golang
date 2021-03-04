package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/internal/db/models"
	"time"
)

type MessageViewRepository struct {
	db *gorm.DB
}

func (r *MessageViewRepository) Init(db *gorm.DB) *MessageViewRepository {
	r.db = db
	return r
}

func (r *MessageViewRepository) UpdateOrCreate(fromID uint, toID uint) error {
	view := &models.MessageView{
		DialogId: toID,
		UserId:   fromID,
	}

	if err := r.db.Where("userId = ? AND dialogId = ?", fromID, toID).FirstOrCreate(&view).Error; err != nil {
		return err
	}

	view.Viewed = time.Now()

	return r.db.Save(&view).Error
}
