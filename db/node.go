package db

import (
	"sync"

	"github.com/muerwre/vault-golang/models"
)

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

func (d DB) GetNodeAlbumRelated(
	ids []uint,
	exclude []uint,
	types string,
	wg *sync.WaitGroup,
	c chan map[string][]models.NodeRelatedItem,
) {
	rows := &[]models.NodeRelatedItem{}
	albums := make(map[string][]models.NodeRelatedItem)

	d.Table("`node_tags_tag` `tags`").
		Select("`tag`.`title` AS `album`, `node`.`thumbnail`, `node`.`id`, `node`.`title`").
		Joins("LEFT JOIN `tag` `tag` ON `tags`.`tagId` = `tag`.`id`").
		Joins("LEFT JOIN `node` `node` ON `tags`.`nodeId` = `node`.`id`").
		Where("`tags`.`tagId` IN (?) AND `node`.`type` = ? AND `node`.`id` NOT IN (?)", []uint(ids), types, exclude).
		Limit(6).
		Scan(&rows)

	for _, v := range *rows {
		albums[v.Album] = append(albums[v.Album], v)
	}

	wg.Done()

	c <- albums
}

func (d DB) GetNodeSimilarRelated(
	ids []uint,
	exclude []uint,
	types string,
	wg *sync.WaitGroup,
	c chan []models.NodeRelatedItem,
) {
	similar := []models.NodeRelatedItem{}

	sq := d.Select("*").
		Table("`node_tags_tag` `tags`").
		Where("`tags`.`tagId` IN (?) AND nodeId NOT IN (?)", []uint(ids), exclude).
		SubQuery()

	d.Table("node").
		Where("`node`.`type` = ?", types).
		Select("`node`.`title`, `node`.`thumbnail`, `node`.`id`, count(t1.nodeId) AS `count`").
		Joins("JOIN (?) as `t1` ON `t1`.`nodeId` = `node`.`id`", sq).
		Order("`count` DESC").
		Group("`t1`.`nodeId`").
		Limit(6).
		Find(&similar)

	wg.Done()

	c <- similar
}

func (d *DB) GetNodeBoris() (node models.Node, err error) {
	d.Where("id = ?", 696).First(&node)

	return node, nil
}
