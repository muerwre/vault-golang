package response

import (
	"github.com/muerwre/vault-golang/db/models"
	"github.com/muerwre/vault-golang/feature/node/response"
)

type LabListResponse struct {
	Nodes []response.FlowResponseNode `json:"nodes"`
	Count int                         `json:"count"`
}

func (r *LabListResponse) Init(nodes []models.Node, count int) *LabListResponse {
	list := make([]response.FlowResponseNode, len(nodes))

	for k, n := range nodes {
		list[k] = *new(response.FlowResponseNode).Init(n)
	}

	r.Nodes = list
	r.Count = count

	return r
}
