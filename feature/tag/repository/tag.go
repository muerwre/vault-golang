package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils"
)

type TagRepository struct {
	db *gorm.DB
}

func (tr *TagRepository) Init(db *gorm.DB) *TagRepository {
	tr.db = db
	return tr
}

func (tr TagRepository) GetByName(name string) (*models.Tag, error) {
	tag := &models.Tag{}
	err := tr.db.Find(&tag, "title = ?", name).Error

	return tag, err
}

func (tr TagRepository) GetNodesOfTag(tag models.Tag, limit int, offset int) ([]*models.Node, int, error) {
	nodes := []*models.Node{}
	t := &models.Tag{}

	if err := tr.db.First(&t, "id = ?", tag.ID).Error; err != nil {
		return nil, 0, err
	}

	q := utils.WhereIsFlowNode(tr.db.Model(&t))
	r := q.Limit(limit).Offset(offset).Order("created_at DESC").Association("Nodes").Find(&nodes)

	if err := r.Error; err != nil {
		return nil, 0, err
	}

	return nodes, q.Association("Nodes").Count(), nil
}

func (tr TagRepository) GetLike(search string) ([]*models.Tag, error) {
	tags := []*models.Tag{}

	if err := tr.db.Limit(25).Find(&tags, "title like concat('%', ?, '%')", search).Error; err != nil {
		return nil, err
	}

	return tags, nil
}

func (tr TagRepository) FindTagsByTitleList(tags []string) ([]*models.Tag, error) {
	r := []*models.Tag{}
	err := tr.db.Where("title IN (?)", tags).Find(&r).Error

	return r, err
}

func (tr TagRepository) CreateTagFromTitle(title string) (*models.Tag, error) {
	tag := &models.Tag{Title: title}
	err := tr.db.Set("gorm:association_autoupdate", false).Save(&tag).Error
	return tag, err
}
