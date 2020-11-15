package usecase

import (
	"fmt"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/db/models"
	repository2 "github.com/muerwre/vault-golang/db/repository"
	"github.com/muerwre/vault-golang/feature/node/constants"
	"github.com/muerwre/vault-golang/feature/node/response"
	validation2 "github.com/muerwre/vault-golang/feature/node/validation"
	fileConstants "github.com/muerwre/vault-golang/feature/upload/constants"
	constants2 "github.com/muerwre/vault-golang/service/notification/constants"
	"github.com/muerwre/vault-golang/service/notification/controller"
	"github.com/muerwre/vault-golang/service/notification/dto"
	"github.com/muerwre/vault-golang/utils/codes"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type NodeUsecase struct {
	notifier controller.NotificationService
	node     repository2.NodeRepository
	file     repository2.FileRepository
	comment  repository2.CommentRepository
	nodeView repository2.NodeViewRepository
}

func (nu *NodeUsecase) Init(db db.DB, notifier controller.NotificationService) *NodeUsecase {
	nu.notifier = notifier
	nu.node = *db.Node
	nu.file = *db.File
	nu.comment = *db.Comment
	nu.nodeView = *db.NodeView
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
		files, err := nu.file.GetByIdList(data.FilesOrder, constants.CommentFileTypes)
		if err != nil {
			return nil, err
		}

		comment.Files = files

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
	nu.file.SetFilesTarget(files, target)
}

func (nu *NodeUsecase) UnsetFilesTarget(files []uint) {
	nu.file.UnsetFilesTarget(files)
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
	if id != 0 {
		comment, err := nu.comment.LoadCommentWithUserAndPhoto(id)

		if err != nil {
			return nil, err
		}

		if *comment.NodeID != node.ID || !comment.CanBeEditedBy(user) {
			return nil, fmt.Errorf(codes.NotEnoughRights)
		}

		return comment, nil
	} else {
		comment := &models.Comment{
			Files: make([]*models.File, 0),
		}

		comment.Node = node
		comment.NodeID = &node.ID
		comment.User = user
		comment.UserID = &user.ID

		return comment, nil
	}
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
	if data.Cover != nil {
		cover, err := nu.file.GetById(data.Cover.ID)
		if err != nil {
			return err
		}

		node.Cover = cover
		node.CoverID = &cover.ID
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
	if data.Model != nil && data.ID != 0 {
		node, err := nu.node.GetNodeWithUser(data.ID)
		if err != nil {
			return nil, err
		}

		if node.Type == "" || !constants.FLOW_NODE_TYPES.Contains(node.Type) {
			return nil, fmt.Errorf(codes.IncorrectType)
		}

		if !node.CanBeEditedBy(u) {
			return nil, fmt.Errorf(codes.NotEnoughRights)
		}

		return node, nil
	} else {
		if data.Type == "" || !constants.FLOW_NODE_TYPES.Contains(data.Type) {
			return nil, fmt.Errorf(codes.IncorrectType)
		}

		node := &models.Node{
			User:       u,
			UserID:     &u.ID,
			Type:       data.Type,
			IsPublic:   true,
			IsPromoted: true,
			Tags:       make([]*models.Tag, 0),
		}

		return node, nil
	}
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
		files, err := nu.file.GetByIdList(data.FilesOrder, constants.NodeFileTypes)

		if err != nil {
			return nil, err
		}

		node.ApplyFiles(files)
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
		_ = nu.node.UpdateDecription(node.ID, comment.Text)
	}
}

func (nu NodeUsecase) UpdateNodeCommentedAt(nid uint) {
	lastComment, _ := nu.node.GetNodeLastComment(nid)

	if lastComment == nil {
		_ = nu.node.UpdateCommentedAt(nid, nil)
	} else {
		_ = nu.node.UpdateCommentedAt(nid, &lastComment.CreatedAt)
	}
}

func (nu NodeUsecase) UpdateNodeSeen(nid uint, uid uint) {
	nu.nodeView.UpdateView(uid, nid)
}

func (nu NodeUsecase) DeleteComment(comment *models.Comment) error {
	return nu.comment.Delete(comment)
}

func (nu NodeUsecase) RestoreComment(comment *models.Comment) error {
	return nu.comment.UnDelete(comment)
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

	node, err := nu.node.GetWithTags(nid)
	if err != nil || len(node.Tags) == 0 {
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
	note := &dto.NotificationDto{
		CreatedAt: node.CreatedAt,
		Timestamp: time.Now(),
		Type:      t,
		ItemId:    node.ID,
	}

	// TODO: add notifier FN for this
	select {
	case nu.notifier.Chan <- note:
		return nil
	default:
		return fmt.Errorf("can't push %s notification, chan closed", t)
	}
}

func (nu NodeUsecase) PushNodeCreateNotification(node models.Node) error {
	return nu.PushNodeNotification(node, constants2.NotifierTypeNodeCreate)
}

func (nu NodeUsecase) PushNodeDeleteNotification(node models.Node) error {
	return nu.PushNodeNotification(node, constants2.NotifierTypeNodeDelete)
}

func (nu NodeUsecase) PushNodeRestoreNotification(node models.Node) error {
	return nu.PushNodeNotification(node, constants2.NotifierTypeNodeRestore)
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
	note := &dto.NotificationDto{
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
	return nu.PushCommentNotification(comment, constants2.NotifierTypeCommentCreate)
}

func (nu NodeUsecase) PushCommentDeleteNotification(comment models.Comment) error {
	return nu.PushCommentNotification(comment, constants2.NotifierTypeCommentDelete)
}

func (nu NodeUsecase) PushCommentRestoreNotification(comment models.Comment) error {
	return nu.PushCommentNotification(comment, constants2.NotifierTypeCommentRestore)
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
		nu.nodeView.UpdateView(uid, node.ID)
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
