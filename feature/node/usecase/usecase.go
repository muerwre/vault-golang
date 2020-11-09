package usecase

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/feature/node/constants"
	"github.com/muerwre/vault-golang/feature/node/repository"
	"github.com/muerwre/vault-golang/feature/node/response"
	validation2 "github.com/muerwre/vault-golang/feature/node/validation"
	fileConstants "github.com/muerwre/vault-golang/feature/upload/constants"
	fileRepository "github.com/muerwre/vault-golang/feature/upload/repository"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/service/notification"
	"github.com/muerwre/vault-golang/utils/codes"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type NodeUsecase struct {
	db       db.DB // TODO: remove, use usecases:
	notifier notification.NotificationService
	node     repository.NodeRepository
	file     fileRepository.FileRepository
}

func (nu *NodeUsecase) Init(db db.DB, notifier notification.NotificationService) *NodeUsecase {
	nu.db = db // TODO: remove, use usecases:
	nu.notifier = notifier
	nu.node = *new(repository.NodeRepository).Init(db.DB)
	nu.file = *new(fileRepository.FileRepository).Init(db.DB)
	return nu
}

func (nu NodeUsecase) UpdateCommentFiles(data *models.Comment, comment *models.Comment) ([]uint, error) {
	// Setting FilesOrder based on sorted Files array of input data
	data.FilesOrder = make(models.CommaUintArray, 0)

	for _, v := range data.Files {
		data.FilesOrder = append(data.FilesOrder, v.ID)
	}

	// Finding out valid comment attaches and sorting them according to files_order
	originFiles := make([]uint, len(comment.FilesOrder))
	copy(originFiles, comment.FilesOrder)

	lostFiles := make(models.CommaUintArray, 0)
	comment.FilesOrder = make(models.CommaUintArray, 0)

	// Loading that files
	if len(data.FilesOrder) > 0 {
		ids, _ := data.FilesOrder.Value()

		comment.Files = make([]*models.File, 0)

		query := nu.db.
			Order(gorm.Expr(fmt.Sprintf("FIELD(id, %s)", ids))).
			Find(
				&comment.Files,
				"id IN (?) AND TYPE IN (?)",
				[]uint(data.FilesOrder),
				constants.CommentFileTypes,
			)

		if query.Error != nil {
			return nil, query.Error
		}

		for i := 0; i < len(comment.Files); i += 1 { // TODO: limit files count
			comment.FilesOrder = append(comment.FilesOrder, comment.Files[i].ID)
		}
	} else {
		comment.Files = make([]*models.File, 0)
		comment.FilesOrder = make(models.CommaUintArray, 0)
	}

	// Detecting lost files
	for _, v := range originFiles {
		if !comment.FilesOrder.Contains(v) {
			lostFiles = append(lostFiles, v)
		}
	}

	return lostFiles, nil
}

func (nu *NodeUsecase) SetFilesTarget(files []uint, target string) {
	if len(files) > 0 {
		nu.db.Model(&models.File{}).Where("id IN (?)", []uint(files)).Update("target", target)
	}
}

func (nu *NodeUsecase) UnsetFilesTarget(files []uint) {
	if len(files) > 0 {
		nu.db.Model(&models.File{}).Where("id IN (?)", files).Update("target", nil)
	}
}

func (nu *NodeUsecase) ValidateAndUpdateCommentText(data *models.Comment, comment *models.Comment) error {
	comment.Text = data.Text

	if len(comment.Text) > constants.MaxCommentLength {
		return fmt.Errorf(codes.CommentTooLong)
	}

	if len(comment.Text) < 1 && len(comment.FilesOrder) == 0 {
		return fmt.Errorf(codes.TextRequired)
	}

	return nil
}

func (nu *NodeUsecase) LoadCommentFromData(id uint, node *models.Node, user *models.User) (*models.Comment, error) {
	comment := &models.Comment{
		Files: make([]*models.File, 0),
	}

	if id != 0 {
		nu.db.Preload("User").Preload("User.Photo").First(&comment, "id = ?", id)
	} else {
		comment.Node = node
		comment.NodeID = &node.ID
		comment.User = user
		comment.UserID = &user.ID
	}

	if *comment.NodeID != node.ID || !comment.CanBeEditedBy(user) {
		return nil, fmt.Errorf(codes.NotEnoughRights)
	}

	return comment, nil
}

func (nu NodeUsecase) UpdateFilesMetadata(data []*models.File, comment []*models.File) {
	for _, df := range data {
		if df == nil || df.Type != fileConstants.FileTypeAudio {
			continue
		}

		for _, cf := range comment {
			if cf != nil && cf.ID == df.ID && cf.Metadata.Title != df.Metadata.Title {
				cf.Metadata.Title = df.Metadata.Title

				if err := nu.file.UpdateMetadata(cf, cf.Metadata); err != nil {
					logrus.Warnf("Can't update file metadata %d: %s", cf.ID, err.Error())
				}

				break
			}
		}
	}
}

func (nu NodeUsecase) UpdateNodeCoverIfChanged(data models.Node, node *models.Node) error {
	// ValidatePatchRequest node cover
	if data.Cover != nil {
		node.Cover = &models.File{}

		query := nu.db.First(node.Cover, "id = ?", data.Cover.ID)

		if query.Error != nil {
			return query.Error
		}

		node.CoverID = &node.Cover.ID
	} else {
		node.Cover = nil
		node.CoverID = nil
	}

	return nil
}

func (nu NodeUsecase) UpdateNodeTitle(data models.Node, node *models.Node) {
	node.Title = data.Title

	if len(node.Title) > constants.MaxNodeTitleLength {
		node.Title = node.Title[:constants.MaxNodeTitleLength]
	}
}

func (nu NodeUsecase) UpdateNodeBlocks(data models.Node, node *models.Node) error {
	node.ApplyBlocks(data.Blocks)

	if val, ok := validation2.NodeValidators[node.Type]; ok {
		err := val(node)

		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf(codes.UnknownNodeType)
	}

	return nil
}

func (nu NodeUsecase) LoadNodeFromData(data models.Node, u *models.User) (*models.Node, error) {
	node := &models.Node{}

	if data.ID != 0 {
		nu.db.Preload("User").Preload("User.Photo").First(&node, "id = ?", data.ID)
		node.Cover = nil
	} else {
		node.User = u
		node.UserID = &u.ID
		node.Type = data.Type
		node.IsPublic = true
		node.IsPromoted = true
		node.Tags = make([]*models.Tag, 0)
	}

	if node.Type == "" || !constants.FLOW_NODE_TYPES.Contains(node.Type) {
		return nil, fmt.Errorf(codes.IncorrectType)
	}

	if !node.CanBeEditedBy(u) {
		return nil, fmt.Errorf(codes.NotEnoughRights)
	}

	return node, nil
}

func (nu NodeUsecase) UpdateNodeFiles(data models.Node, node *models.Node) ([]uint, error) {
	// Finding out valid comment attaches and sorting them according to files_order
	originFiles := make([]uint, len(node.FilesOrder))
	copy(originFiles, node.FilesOrder)

	// Setting FilesOrder based on sorted Files array of input data
	data.FilesOrder = make(models.CommaUintArray, 0)

	for _, v := range data.Files {
		if v == nil {
			continue
		}

		data.FilesOrder = append(data.FilesOrder, v.ID)
	}

	if len(data.FilesOrder) > 0 {
		ids, _ := data.FilesOrder.Value()

		data.Files = make([]*models.File, 0)

		query := nu.db.
			Order(gorm.Expr(fmt.Sprintf("FIELD(id, %s)", ids))).
			Find(&data.Files, "id IN (?)", []uint(data.FilesOrder))

		if query.Error != nil {
			return nil, query.Error
		}

		node.ApplyFiles(data.Files)
	} else {
		node.Files = make([]*models.File, 0)
		node.FilesOrder = make(models.CommaUintArray, 0)
	}

	// Detecting lost files
	lostFiles := make(models.CommaUintArray, 0)

	for _, v := range originFiles {
		if !node.FilesOrder.Contains(v) {
			lostFiles = append(lostFiles, v)
		}
	}

	return lostFiles, nil
}

func (nu NodeUsecase) UnsetNodeCoverTarget(data models.Node, node *models.Node) {
	if node.Cover != nil && (data.Cover == nil || data.Cover.ID != node.Cover.ID) {
		nu.UnsetFilesTarget([]uint{node.Cover.ID})
	}
}

func (nu NodeUsecase) UpdateBriefFromComment(node *models.Node, comment *models.Comment) {
	if node.Description == "" && *comment.UserID == *node.UserID && len(comment.Text) >= 64 {
		node.Description = comment.Text
		nu.db.Model(&models.Node{}).Where("id = ?", node.ID).Update("description", comment.Text)
	}
}

func (nu NodeUsecase) UpdateNodeCommentedAt(nid uint) {
	lastComment, _ := nu.node.GetNodeLastComment(nid)

	if lastComment == nil {
		nu.db.Model(&models.Node{}).Where("id = ?", nid).Update("commented_at", nil)
	} else {
		nu.db.Model(&models.Node{}).Where("id = ?", nid).Update("commented_at", lastComment.CreatedAt)
	}
}

func (nu NodeUsecase) UpdateNodeSeen(nid uint, uid uint) {
	nu.db.NodeView.UpdateView(uid, nid)
}

func (nu NodeUsecase) DeleteComment(comment *models.Comment) error {
	return nu.db.Delete(&comment).Error
}

func (nu NodeUsecase) RestoreComment(comment *models.Comment) error {
	comment.DeletedAt = nil

	return nu.db.Model(&comment).
		Unscoped().
		Where("id = ?", comment.ID).
		Update("deletedAt", nil).Error
}

func (nu NodeUsecase) SeparateNodeTags(tags []*models.Tag) ([]uint, []uint) {
	var similar []uint
	var album []uint

	for _, v := range tags {
		if v.Title[:1] == "/" {
			album = append(album, v.ID)
		} else {
			similar = append(similar, v.ID)
		}
	}

	return similar, album
}

func (nu NodeUsecase) GetNodeRelated(nid uint) (*response.NodeRelatedResponse, error) {
	related := &response.NodeRelatedResponse{
		Albums:  map[string][]models.NodeRelatedItem{},
		Similar: []models.NodeRelatedItem{},
	}

	node := &models.Node{}
	if err := nu.db.Preload("Tags").First(&node, "id = ?", nid).Error; err != nil || len(node.Tags) == 0 {
		return related, nil
	}

	similarIds, albumIds := nu.SeparateNodeTags(node.Tags)

	var wg sync.WaitGroup
	wg.Add(2)

	albumsChan := make(chan map[string][]models.NodeRelatedItem)
	similarChan := make(chan []models.NodeRelatedItem)

	go nu.node.GetNodeAlbumRelated(albumIds, []uint{node.ID}, node.Type, &wg, albumsChan)
	go nu.node.GetNodeSimilarRelated(similarIds, []uint{node.ID}, node.Type, &wg, similarChan)

	wg.Wait()

	related.Albums = <-albumsChan
	related.Similar = <-similarChan

	return related, nil
}

func (nu NodeUsecase) PushNodeNotification(node models.Node, t string) error {
	note := &notification.NotifierItem{
		CreatedAt: node.CreatedAt,
		Timestamp: time.Now(),
		Type:      t,
		ItemId:    node.ID,
	}

	select {
	case nu.notifier.Chan <- note:
		return nil
	default:
		return fmt.Errorf("Can't push %s notification, chan closed", t)
	}
}

func (nu NodeUsecase) PushNodeCreateNotification(node models.Node) error {
	return nu.PushNodeNotification(node, notification.NotifierTypeNodeCreate)
}

func (nu NodeUsecase) PushNodeDeleteNotification(node models.Node) error {
	return nu.PushNodeNotification(node, notification.NotifierTypeNodeDelete)
}

func (nu NodeUsecase) PushNodeRestoreNotification(node models.Node) error {
	return nu.PushNodeNotification(node, notification.NotifierTypeNodeRestore)
}

func (nu NodeUsecase) PushNodeCreateNotificationIfNeeded(data models.Node, node models.Node) error {
	switch {
	case data.ID == 0 && node.ID != 0 && node.IsFlowType() && node.IsPublic:
		return nu.PushNodeCreateNotification(node)
	default:
		return nil
	}
}

func (nu NodeUsecase) PushCommentCreateNotificationIfNeeded(data models.Comment, comment models.Comment) error {
	switch {
	case data.ID == 0 && comment.ID != 0 && comment.Node != nil && comment.Node.IsFlowType() && comment.Node.IsPublic:
		return nu.PushCommentCreateNotification(comment)
	default:
		return nil
	}
}

func (nu NodeUsecase) PushCommentNotification(comment models.Comment, t string) error {
	note := &notification.NotifierItem{
		CreatedAt: comment.CreatedAt,
		Timestamp: time.Now(),
		Type:      t,
		ItemId:    comment.ID,
	}

	select {
	case nu.notifier.Chan <- note:
		return nil
	default:
		return fmt.Errorf("Can't push %s notification, chan closed", t)
	}
}

func (nu NodeUsecase) PushCommentCreateNotification(comment models.Comment) error {
	return nu.PushCommentNotification(comment, notification.NotifierTypeCommentCreate)
}

func (nu NodeUsecase) PushCommentDeleteNotification(comment models.Comment) error {
	return nu.PushCommentNotification(comment, notification.NotifierTypeCommentDelete)
}

func (nu NodeUsecase) PushCommentRestoreNotification(comment models.Comment) error {
	return nu.PushCommentNotification(comment, notification.NotifierTypeCommentRestore)
}

func (nu NodeUsecase) GetNodeWithLikesAndFiles(id int, role string, uid uint) (*models.Node, error) {
	node, err := nu.node.GetFullNode(
		id,
		role == models.USER_ROLES.ADMIN,
		uid,
	)

	if err != nil {
		return nil, err
	}

	if uid != 0 {
		node.IsLiked = nu.node.IsNodeLikedBy(node, uid)
		nu.db.NodeView.UpdateView(uid, node.ID)
	}

	node.LikeCount = nu.node.GetNodeLikeCount(node)
	node.Files, _ = nu.file.GetFilesByIds([]uint(node.FilesOrder))

	node.SortFiles()

	return node, nil
}

func (nu NodeUsecase) GetComments(id uint, take int, skip int, order string) ([]*models.Comment, int, error) {
	if take <= 0 {
		take = 100
	}

	if skip < 0 {
		skip = 0
	}

	if order != "ASC" {
		order = "DESC"
	}

	return nu.node.GetComments(id, take, skip, order)
}

func (nu NodeUsecase) GetDiffNodesBefore(start *time.Time) ([]models.Node, error) {
	return nu.node.GetDiffNodesBefore(start)
}

func (nu NodeUsecase) GetDiffNodesAfter(end *time.Time, take uint) ([]models.Node, error) {
	if take <= 0 {
		take = 50
	}

	return nu.node.GetDiffNodesAfter(end, take)
}

func (nu NodeUsecase) GetDiffHeroes(shouldSearch bool) ([]models.Node, error) {
	if !shouldSearch {
		return make([]models.Node, 0), nil
	}

	return nu.node.GetDiffHeroes()
}

func (nu NodeUsecase) GetDiffUpdated(uid uint, shouldSearch bool) ([]models.Node, []uint, error) {
	if !shouldSearch {
		return make([]models.Node, 0), make([]uint, 0), nil
	}

	updated, err := nu.node.GetDiffUpdated(uid, 10)
	if err != nil {
		return nil, nil, err
	}

	exclude := make([]uint, len(updated)+1)
	exclude[0] = 0

	for k, v := range updated {
		exclude[k+1] = v.ID
	}

	return updated, exclude, nil
}

func (nu NodeUsecase) GetDiffRecent(exclude []uint, shouldSearch bool) ([]models.Node, error) {
	if !shouldSearch {
		return make([]models.Node, 0), nil
	}

	return nu.node.GetDiffRecent(16, exclude)
}

func (nu NodeUsecase) GetDiffValid(start *time.Time, end *time.Time, shouldSearch bool) ([]uint, error) {
	if !shouldSearch {
		return make([]uint, 0), nil
	}

	return nu.node.GetDiffValid(start, end)
}

func (nu NodeUsecase) GetDeletedComment(cid uint, nid uint, u models.User) (*models.Comment, error) {
	comment, err := nu.node.GetCommentByIdWithDeleted(cid)

	if err != nil {
		return nil, err
	}

	if *comment.NodeID != nid {
		return nil, fmt.Errorf("comment %d is not of node %d", cid, nid)
	}

	if !u.CanEditComment(comment) {
		return nil, fmt.Errorf("user %s can't edit comment %d", u.Username, cid)
	}

	return comment, err
}

func (nu NodeUsecase) GetCommentableNodeById(nid uint) (*models.Node, error) {
	node, err := nu.node.GetById(nid)

	if err != nil {
		return nil, err
	}

	if !node.CanBeCommented() {
		return nil, fmt.Errorf("node [%d](%s) can't be commented, but someone is trying", nid, node.Title)
	}

	return node, nil
}

func (nu NodeUsecase) SaveCommentWithFiles(comment *models.Comment) error {
	return nu.node.SaveCommentWithFiles(comment)
}

func (nu NodeUsecase) GetTaggableNodeById(nid uint, u *models.User) (*models.Node, error) {
	node, err := nu.node.GetById(nid)

	if err != nil {
		return nil, err
	}

	if !node.CanBeTaggedBy(u) {
		return nil, fmt.Errorf("node [%d](%s) can't be commented, but someone is trying", nid, node.Title)
	}

	return node, nil
}

func (nu NodeUsecase) UpdateNodeTags(node *models.Node, tags []*models.Tag) error {
	return nu.node.UpdateNodeTags(node, tags)
}

func (nu NodeUsecase) UpdateNodeLikeByUser(node *models.Node, user *models.User, isLiked bool) error {
	switch isLiked {
	case true:
		return nu.node.LikeNode(node, user)
	default:
		return nu.node.DislikeNode(node, user)
	}
}

func (nu NodeUsecase) GetDeletedNode(nid uint, u *models.User) (*models.Node, error) {
	node, err := nu.node.GetDeletedNode(nid)
	if err != nil {
		return nil, err
	}

	if !node.CanBeEditedBy(u) {
		return nil, fmt.Errorf("Node %d can't be edited by user %s", nid, u.Username)
	}

	return node, nil
}

func (nu NodeUsecase) UpdateNodeLocked(node *models.Node, isLocked bool) error {
	switch isLocked {
	case true:
		if err := nu.node.LockNode(node); err != nil {
			return err
		}

		_ = nu.PushNodeDeleteNotification(*node)
	default:
		if err := nu.node.UnlockNode(node); err != nil {
			return err
		}

		_ = nu.PushNodeRestoreNotification(*node)
	}

	return nil
}

func (nu NodeUsecase) GetHeroeableNodeById(nid uint, u *models.User) (*models.Node, error) {
	node, err := nu.node.GetById(nid)
	if err != nil {
		return nil, err
	}

	if !node.CanBeHeroedBy(u) {
		return nil, fmt.Errorf("node %d can't be heroed by user %s", u.Username)
	}

	return node, nil
}

func (nu NodeUsecase) UpdateNodeIsHeroic(node *models.Node, isHeroic bool) error {
	return nu.node.UpdateNodeIsHeroic(node, isHeroic)
}

func (nu NodeUsecase) GetEditableNodeById(nid uint, u *models.User) (*models.Node, error) {
	node, err := nu.node.GetById(nid)
	if err != nil {
		return nil, err
	}

	if !node.CanBeEditedBy(u) {
		return nil, fmt.Errorf("node %d can't be edited by user %s", u.Username)
	}

	return node, nil
}

func (nu NodeUsecase) UpdateNodeFlow(node *models.Node, flow models.NodeFlow) error {
	return nu.node.UpdateNodeFlow(node, flow)
}

func (nu NodeUsecase) SaveNodeWithFiles(node *models.Node) error {
	return nu.node.SaveNodeWithFiles(node)
}
