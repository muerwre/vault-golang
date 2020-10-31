package repository

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/constants"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils"
	"github.com/muerwre/vault-golang/utils/codes"
	"sync"
)

type NodeRepository struct {
	db *gorm.DB
}

func (nr *NodeRepository) Init(db *gorm.DB) *NodeRepository {
	nr.db = db
	return nr
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
	albums := make(map[string][]models.NodeRelatedItem, 0)

	d.Table("`node_tags_tag` `tags`").
		Select("`tag`.`title` AS `album`, `node`.`thumbnail`, `node`.`id`, `node`.`title`").
		Joins("LEFT JOIN `tag` `tag` ON `tags`.`tagId` = `tag`.`id`").
		Joins("LEFT JOIN `node` `node` ON `tags`.`nodeId` = `node`.`id`").
		Where("`tags`.`tagId` IN (?) AND `node`.`type` = ? AND `node`.`id` NOT IN (?) AND `node`.`deleted_at` IS NULL", ids, types, exclude).
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
	similar := []models.NodeRelatedItem{}

	sq := nr.db.Select("*").
		Table("`node_tags_tag` `tags`").
		Where("`tags`.`tagId` IN (?) AND nodeId NOT IN (?)", []uint(ids), exclude).
		SubQuery()

	nr.db.Table("node").
		Where("`node`.`type` = ? AND `node`.`deleted_at` IS NULL", types).
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
	nr.db.Where("id = ?", constants.BorisNodeId).First(&node)

	return node, nil
}

func (nr NodeRepository) GetNodeTypeCount(t string) int {
	count := 0
	nr.db.Model(&models.Node{}).Where("node.type = ?", t).Count(&count)
	return count
}

func (nr NodeRepository) GetImagesCount() int {
	return nr.GetNodeTypeCount(constants.NODE_TYPES.IMAGE)
}

func (nr NodeRepository) GetAudiosCount() int {
	return nr.GetNodeTypeCount(constants.NODE_TYPES.AUDIO)
}

func (nr NodeRepository) GetVideosCount() int {
	return nr.GetNodeTypeCount(constants.NODE_TYPES.VIDEO)
}

func (nr NodeRepository) GetTextsCount() int {
	return nr.GetNodeTypeCount(constants.NODE_TYPES.TEXT)
}

func (nr NodeRepository) GetCommentsCount() (count int) {
	nr.db.Model(&models.Comment{}).Count(&count)
	return
}

func (nr NodeRepository) GetFlowLastPost() (*models.Node, error) {
	node := &models.Node{}

	utils.WhereIsFlowNode(nr.db.Model(&node).Order("created_at DESC").Limit(1)).First(&node)

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

func (nr NodeRepository) GetComments(id int, take int, skip int, order string) (*[]*models.Comment, int) {
	comments := &[]*models.Comment{}
	count := 0

	q := nr.db.
		Where("nodeId = ?", id).
		Order(fmt.Sprintf("created_at %s", order))

	q.Model(&comments).Count(&count)

	q.Preload("User").
		Preload("Files").
		Preload("User.Photo").
		Offset(skip).
		Limit(take).
		Find(&comments)

	for _, v := range *comments {
		v.SortFiles()
	}

	return comments, count
}

func (nr NodeRepository) GetForSearch(
	text string,
	take int,
	skip int,
) ([]*models.Node, int) {
	count := 0
	res := make([]*models.Node, 0)

	query := nr.db.
		Where("(title like concat('%', ?, '%') OR description like concat('%', ?, '%'))", text, text).
		Order(fmt.Sprintf("title LIKE concat('%s', '%%') DESC", text)).
		Order(fmt.Sprintf("description LIKE concat('%s', '%%') DESC", text)).
		Order("created_at DESC ")

	query = utils.WhereIsFlowNode(query)

	query.Model(&models.Node{}).Count(&count)
	query.Limit(take).Offset(skip).Find(&res)

	return res, count
}

func (nr NodeRepository) GetById(id uint) (*models.Node, error) {
	node := &models.Node{}
	query := nr.db.First(&node, "id = ?", id)

	return node, query.Error
}

func (nr NodeRepository) SaveCommentWithFiles(comment *models.Comment) error {
	query := nr.db.
		Set("gorm:association_autoupdate", false).
		Set("gorm:association_save_reference", false).
		Save(&comment).
		Association("Files").
		Replace(comment.Files)

	if query.Error != nil {
		return query.Error
	}

	if len(comment.FilesOrder) > 0 {
		nr.db.Model(&models.File{}).
			Where("id IN (?)", []uint(comment.FilesOrder)).
			Update("Target", "comment")
	} else {
		// TODO: remove this after moving to CommentResponse
		comment.Files = make([]*models.File, 0) // pass empty array to response
	}

	return nil
}

func (nr NodeRepository) SaveNodeWithFiles(node *models.Node) error {
	// Save node and its files
	query := nr.db.
		Set("gorm:association_autoupdate", false).
		Set("gorm:association_save_reference", false).
		Save(&node).
		Association("Files").
		Replace(node.Files)

	if query.Error != nil {
		return query.Error
	}

	if len(node.FilesOrder) > 0 {
		nr.db.Model(&models.File{}).
			Where("id IN (?)", []uint(node.FilesOrder)).
			Update("Target", "comment")
	} else {
		// TODO: remove this after moving to CommentResponse
		node.Files = make([]*models.File, 0) // pass empty array to response
	}

	return nil
}

// GetCommentByIdWithDeleted gets comment by cid with deleted ones
func (nr NodeRepository) GetCommentByIdWithDeleted(cid uint) (*models.Comment, error) {
	comment := &models.Comment{}
	q := nr.db.Unscoped().Where("id = ?", cid).First(&comment)
	return comment, q.Error
}

func (nr NodeRepository) GetNodeLastComment(nid uint) (*models.Comment, error) {
	comment := &models.Comment{}

	if err := nr.db.Model(&comment).Where("nodeID = ?", nid).Order("created_at DESC").First(&comment).Error; err != nil {
		return nil, err
	}

	return comment, nil
}
