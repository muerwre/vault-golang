package response

import (
	"github.com/muerwre/vault-golang/db/models"
	"time"
)

type SearchNodeResponse struct {
	Total int                      `json:"total"`
	Nodes []SearchNodeResponseNode `json:"nodes"`
}

type SearchNodeResponseNode struct {
	Id        uint      `json:"id"`
	Thumbnail string    `json:"thumbnail"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

func (n *SearchNodeResponseNode) Init(node models.Node) *SearchNodeResponseNode {
	n.Id = node.ID
	n.Thumbnail = node.Thumbnail
	n.Title = node.Title
	n.CreatedAt = node.CreatedAt

	return n
}
