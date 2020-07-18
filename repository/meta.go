package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/models"
)

type MetaRepository struct {
	db *gorm.DB
}

func (mr *MetaRepository) Init(db *gorm.DB) *MetaRepository {
	mr.db = db
	return mr
}

func (mr *MetaRepository) GetEmbedsById(addresses []string, provider string) (map[string]*models.Embed, error) {
	result := make([]models.Embed, 0)
	mr.db.Model(&result).Where("provider = ? AND address in (?)", provider, addresses).Find(&result)

	withIds := make(map[string]*models.Embed, len(result))

	for _, v := range result {
		withIds[v.Address] = &v
	}

	return withIds, nil
}

func (mr *MetaRepository) SaveEmbeds(items []models.Embed) error {
	ts := mr.db.Begin()
	for _, v := range items {
		ts.Create(&v)
	}
	ts.Commit()
	return nil
}
