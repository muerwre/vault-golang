package repository

import (
	"fmt"
	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
	"sync"
)

type NodeRepository struct {
	db *gorm.DB
}

func (nr *NodeRepository) Init(db *gorm.DB) {
	nr.db = db
}

func (nr *NodeRepository) WhereIsFlowNode(d *gorm.DB) *gorm.DB {
	return d.Where(
		"deleted_at IS NULL AND is_promoted = 1 AND is_public = 1 AND type IN (?)",
		structs.Values(models.FLOW_NODE_TYPES),
	)
}

func (nr NodeRepository) IsNodeLikedBy(node *models.Node, uid uint) bool {
	c := 0

	nr.db.Table("like").Where("nodeId = ? AND userId = ?", node.ID, uid).Count(&c)

	return c > 0
}

func (nr NodeRepository) GetNodeLikeCount(node *models.Node) (c int) {
	nr.db.Table("like").Where("nodeId = ?", node.ID).Count(&c)
	return
}

func (nr NodeRepository) GetNodeAlbumRelated(
	ids []uint,
	exclude []uint,
	types string,
	wg *sync.WaitGroup,
	c chan map[string][]models.NodeRelatedItem,
) {
	d := nr.db
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

func (nr NodeRepository) GetNodeSimilarRelated(
	ids []uint,
	exclude []uint,
	types string,
	wg *sync.WaitGroup,
	c chan []models.NodeRelatedItem,
) {
	d := nr.db
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

func (nr NodeRepository) GetNodeBoris() (node models.Node, err error) {
	nr.db.Where("id = ?", 696).First(&node)

	return node, nil
}

func (nr NodeRepository) GetImagesCount() (count int) {
	nr.db.Model(&models.Node{}).Where("node.type = ?", models.NODE_TYPES.IMAGE).Count(&count)
	return
}

func (nr NodeRepository) GetAudiosCount() (count int) {
	nr.db.Model(&models.Node{}).Where("node.type = ?", models.NODE_TYPES.AUDIO).Count(&count)
	return
}

func (nr NodeRepository) GetVideosCount() (count int) {
	nr.db.Model(&models.Node{}).Where("node.type = ?", models.NODE_TYPES.VIDEO).Count(&count)
	return
}

func (nr NodeRepository) GetTextsCount() (count int) {
	nr.db.Model(&models.Node{}).Where("node.type = ?", models.NODE_TYPES.TEXT).Count(&count)
	return
}

func (nr NodeRepository) GetCommentsCount() (count int) {
	nr.db.Model(&models.Comment{}).Count(&count)
	return
}

func (nr NodeRepository) GetFlowLastPost() (*models.Node, error) {
	node := &models.Node{}

	nr.WhereIsFlowNode(nr.db.Model(&node).Order("created_at DESC").Limit(1)).First(&node)

	if node.ID == 0 {
		return nil, fmt.Errorf(codes.NodeNotFound)
	}

	return node, nil
}

func (nr NodeRepository) GetFullNode(id int, isAdmin bool, uid uint) (*models.Node, error) {
	node := &models.Node{}

	q := nr.db.Unscoped().
		Preload("Tags").
		Preload("User").
		Preload("Cover")

	if uid != 0 && isAdmin {
		q.First(&node, "id = ?", id)
	} else {
		q.First(&node, "id = ? AND (deleted_at IS NULL OR userID = ?)", id, uid)
	}

	if node.ID == 0 {
		return nil, fmt.Errorf(codes.NodeNotFound)
	}

	return node, nil
}

func (nr NodeRepository) GetComments(id int, take int, skip int, order string) *[]*models.Comment {
	comments := &[]*models.Comment{}

	nr.db.Preload("User").
		Preload("Files").
		Preload("User.Photo").
		Where("nodeId = ?", id).
		Offset(skip).
		Limit(take).
		Order(fmt.Sprintf("created_at %s", order)).
		Find(&comments)

	for _, v := range *comments {
		v.SortFiles()
	}

	return comments
}
