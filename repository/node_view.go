package repository

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/models"
)

type NodeViewRepository struct {
	db *gorm.DB
}

func (nvr *NodeViewRepository) Init(db *gorm.DB) *NodeViewRepository {
	nvr.db = db
	return nvr
}

func (nvr *NodeViewRepository) GetOne(uid uint, nid uint) (*models.NodeView, error) {
	view := &models.NodeView{}
	nvr.db.Model(&view).Where("userId = ? AND nodeId = ?", uid, nid).First(&view)

	if view.ID == 0 {
		return nil, fmt.Errorf("can't load node view for (nodeId = %d, userId = %d)", nid, uid)
	}

	return view, nil
}
