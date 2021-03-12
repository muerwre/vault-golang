package response

import (
	"github.com/muerwre/vault-golang/internal/db/models"
)

type LabListResponse struct {
	Nodes []models.Node `json:"nodes"`
	Count int           `json:"count"`
}

func (r *LabListResponse) Init(nodes []models.Node, count int) *LabListResponse {
	// TODO: make map to LabNodeResponse

	r.Nodes = nodes
	r.Count = count

	return r
}
