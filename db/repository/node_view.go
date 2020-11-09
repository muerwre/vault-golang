package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/db/models"
	"time"
)

type NodeViewRepository struct {
	db *gorm.DB
}

func (r *NodeViewRepository) Init(db *gorm.DB) *NodeViewRepository {
	r.db = db
	return r
}

func (r *NodeViewRepository) GetOrCreateOne(uid uint, nid uint) (*models.NodeView, error) {
	view := &models.NodeView{
		UserID: uid,
		NodeID: nid,
	}

	if err := r.db.Model(&view).Where("userId = ? AND nodeId = ?", uid, nid).FirstOrCreate(&view).Error; err != nil {
		return nil, err
	}

	return view, nil
}

func (r *NodeViewRepository) UpdateView(uid uint, nid uint) *models.NodeView {
	nv := &models.NodeView{
		NodeID:  nid,
		UserID:  uid,
		Visited: time.Now(),
	}

	r.db.Model(&nv).FirstOrCreate(&nv, "nodeId = ? AND userId = ?", nid, uid)
	r.db.Model(&nv).Update("visited", time.Now())

	return nv
}
