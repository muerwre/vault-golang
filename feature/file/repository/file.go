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

	defer rows.Close()
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

func (fr FileRepository) Save(f *models.File) error {
	return fr.db.Save(&f).Error
}

func (fr FileRepository) UpdateMetadata(f *models.File, m models.FileMetadata) error {
	return fr.db.Model(&f).Update("metadata", m).Error
}

func (fr FileRepository) GetById(id uint) (*models.File, error) {
	file := &models.File{}
	query := fr.db.First(&file, "id = ?", id)

	return file, query.Error
}