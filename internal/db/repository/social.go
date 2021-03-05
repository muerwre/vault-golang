package repository

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/internal/db/models"
	"github.com/muerwre/vault-golang/pkg/codes"
	"github.com/sirupsen/logrus"
)

type OauthRepository struct {
	db *gorm.DB
}

func (sr *OauthRepository) Init(db *gorm.DB) *OauthRepository {
	sr.db = db
	return sr
}

func (sr *OauthRepository) FindOne(provider string, id string) (*models.Social, error) {
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

func (sr *OauthRepository) Create(social *models.Social) error {
	return sr.db.Create(&social).Error
}

func (sr *OauthRepository) OfUser(id uint) ([]*models.Social, error) {
	list := make([]*models.Social, 0)
	sr.db.Model(&list).Where("userId = ?", id).Scan(&list)
	return list, nil
}

func (sr *OauthRepository) DeleteOfUser(uid uint, provider string, id string) error {
	sr.db.Delete(&models.Social{}, "userId = ? AND provider = ? AND account_id = ?", uid, provider, id)
	return nil
}
