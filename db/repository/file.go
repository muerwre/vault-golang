package repository

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/db/models"
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

func (fr FileRepository) GetByIdList(order models.CommaUintArray, types []string) ([]*models.File, error) {
	ids, _ := order.Value()
	files := make([]*models.File, 0)

	err := fr.db.
		Order(gorm.Expr(fmt.Sprintf("FIELD(id, %s)", ids))).
		Find(
			&files,
			"id IN (?) AND TYPE IN (?)",
			[]uint(order),
			types,
		).Error

	if err != nil {
		return nil, err
	}

	return files, nil
}

func (fr FileRepository) SetFilesTarget(files []uint, target string) {
	if len(files) > 0 {
		fr.db.Model(&models.File{}).Where("id IN (?)", files).Update("target", target)
	}
}

func (fr FileRepository) UnsetFilesTarget(files []uint) {
	if len(files) > 0 {
		fr.db.Model(&models.File{}).Where("id IN (?)", files).Update("target", nil)
	}
}
