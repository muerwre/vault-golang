package repository

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
	"github.com/sirupsen/logrus"
)

type SocialRepository struct {
	db *gorm.DB
}

func (sr *SocialRepository) Init(db *gorm.DB) *SocialRepository {
	sr.db = db
	return sr
}

func (sr *SocialRepository) FindOne(provider string, id string) (*models.Social, error) {
	social := &models.Social{}

	err := sr.db.
		Model(&social).
		Where("provider = ? AND account_id = ?", provider, id).
		Preload("User").
		First(&social).Error

	if social.ID == 0 || err != nil {
		logrus.Infof("Can't get social for user", err.Error())
		return nil, fmt.Errorf(codes.UserNotFound)
	}

	return social, nil
}

func (sr *SocialRepository) Create(social *models.Social) {
	sr.db.Create(&social)
}

func (sr *SocialRepository) OfUser(id uint) ([]*models.Social, error) {
	list := make([]*models.Social, 0)
	sr.db.Model(&list).Where("userId = ?", id).Scan(&list)
	return list, nil
}

func (sr *SocialRepository) DeleteOfUser(uid uint, provider string, id string) error {
	sr.db.Delete(&models.Social{}, "userId = ? AND provider = ? AND account_id = ?", uid, provider, id)
	return nil
}
