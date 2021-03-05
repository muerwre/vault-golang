package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/internal/db/models"
	"time"
)

type CommentRepository struct {
	db *gorm.DB
}

func (cr *CommentRepository) Init(db *gorm.DB) *CommentRepository {
	cr.db = db
	return cr
}

func (cr CommentRepository) LoadCommentWithUserAndPhoto(id uint) (*models.Comment, error) {
	comment := &models.Comment{
		Files: make([]*models.File, 0),
	}

	if err := cr.db.Preload("User").Preload("User.Photo").First(&comment, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return comment, nil
}

func (cr CommentRepository) Delete(comment *models.Comment) error {
	now := time.Now()
	comment.DeletedAt = &now
	return cr.db.Delete(comment).Error
}

func (cr CommentRepository) UnDelete(comment *models.Comment) error {
	comment.DeletedAt = nil
	return cr.db.Model(comment).Unscoped().
		Where("id = ?", comment.ID).
		Update("deletedAt", nil).Error
}
