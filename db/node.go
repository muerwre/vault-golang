package db

import "github.com/muerwre/vault-golang/models"

func (d *DB) GetNodeFiles(nid uint) ([]*models.File, error) {
	return nil, nil
}

func (d *DB) IsNodeLikedBy(node *models.Node, uid uint) bool {
	c := 0

	d.Table("like").Where("nodeId = ? AND userId = ?", node.ID, uid).Count(&c)

	return c > 0
}

func (d *DB) GetNodeLikeCount(node *models.Node) (c int) {
	d.Table("like").Where("nodeId = ?", node.ID).Count(&c)
	return
}
