package usecase

import (
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/db/models"
	"github.com/muerwre/vault-golang/db/repository"
	"time"
)

type LabUsecase struct {
	node repository.NodeRepository
}

func (uc *LabUsecase) Init(db db.DB) *LabUsecase {
	uc.node = *db.Node
	return uc
}

func (uc LabUsecase) GetList(after *time.Time, limit int) ([]models.Node, int, error) {
	a := time.Now()
	if after != nil {
		a = *after
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	return uc.node.GetLabNodes(a, limit)
}
