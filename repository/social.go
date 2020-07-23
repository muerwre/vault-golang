package repository

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
)

type SocialRepository struct {
	db *gorm.DB
}

func (sr *SocialRepository) Init(db *gorm.DB) *SocialRepository {
	sr.db = db
	return sr
}

func (sr *SocialRepository) FindOne(provider string, id string) (social *models.Social, err error) {
	social = &models.Social{}
	sr.db.
		Where("provider = ? AND account_id = ?", provider, id).
		Preload("User").
		First(&social)

	if social.ID == 0 {
		return nil, fmt.Errorf(codes.UserNotFound)
	}

	return
}

func (sr *SocialRepository) Create(social *models.Social) {
	sr.db.Create(&social)
}