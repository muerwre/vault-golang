package repository

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/models"
	"time"
)

type NodeViewRepository struct {
	db *gorm.DB
}

func (r *NodeViewRepository) Init(db *gorm.DB) *NodeViewRepository {
	r.db = db
	return r
}

func (r *NodeViewRepository) GetOne(uid uint, nid uint) (*models.NodeView, error) {
	view := &models.NodeView{}
	r.db.Model(&view).Where("userId = ? AND nodeId = ?", uid, nid).First(&view)

	if view.ID == 0 {
		return nil, fmt.Errorf("can't load node view for (nodeId = %d, userId = %d)", nid, uid)
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
