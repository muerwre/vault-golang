package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/models"
)

type FileRepository struct {
	db *gorm.DB
}

func (fr *FileRepository) Init(db *gorm.DB) *FileRepository {
	fr.db = db
	return fr
}

func (fr FileRepository) GetTotalCount() (count int) {
	fr.db.Model(&models.File{}).Count(&count)
	return
}

func (fr FileRepository) GetTotalSize() (count int) {
	rows, err := fr.db.Model(&models.File{}).
		Select("SUM(size) as size").
		Where("target IS NOT NULL").
		Rows()

	if err != nil {
		return 0
	}

	rows.Next()
	err = rows.Scan(&count)

	if err != nil {
		return 0
	}

	return
}

func (fr FileRepository) GetFilesByIds(ids []uint) ([]*models.File, error) {
	files := make([]*models.File, len(ids))

	fr.db.Where("id IN (?)", ids).Find(&files)

	return files, nil
}

func (fr FileRepository) Save(f *models.File) {
	fr.db.Save(&f)
}

func (fr FileRepository) UpdateTargetForIds(ids []uint, target string) {
	if len(ids) > 0 {
		fr.db.Model(&models.File{}).Where("id IN (?)", ids).Update("Target", target)
	}
}
