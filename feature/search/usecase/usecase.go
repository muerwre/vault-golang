package usecase

import (
	"github.com/muerwre/vault-golang/db"
	repository2 "github.com/muerwre/vault-golang/db/repository"
	"github.com/muerwre/vault-golang/feature/search/request"
	"github.com/muerwre/vault-golang/feature/search/response"
)

type SearchUsecase struct {
	search repository2.SearchRepository
	node   repository2.NodeRepository
}

func (su *SearchUsecase) Init(db db.DB) *SearchUsecase {
	su.search = *db.Search
	su.node = *db.Node
	return su
}

func (su SearchUsecase) GetNodesForSearch(req request.SearchNodeRequest) *response.SearchNodeResponse {
	resp := &response.SearchNodeResponse{
		Nodes: make([]response.SearchNodeResponseNode, 0),
	}
	if len(req.Text) == 0 {
		return resp
	}

	nodes, count := su.node.GetForSearch(req.Text, req.Take, req.Skip)
	for _, v := range nodes {
		node := new(response.SearchNodeResponseNode).Init(*v)
		resp.Nodes = append(resp.Nodes, *node)
	}

	resp.Total = count

	return resp
}
