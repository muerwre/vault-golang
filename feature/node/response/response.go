package response

import (
	"github.com/muerwre/vault-golang/db/models"
	"time"
)

type NodeRelatedResponse struct {
	Albums  map[string][]models.NodeRelatedItem `json:"albums"`
	Similar []models.NodeRelatedItem            `json:"similar"`
}

type FlowNodeResponseUser struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Photo    string `json:"photo"`
}

type FlowResponse struct {
	Before  []FlowResponseNode `json:"before"`
	After   []FlowResponseNode `json:"after"`
	Heroes  []FlowResponseNode `json:"heroes"`
	Updated []FlowResponseNode `json:"updated"`
	Recent  []FlowResponseNode `json:"recent"`
	Valid   []uint             `json:"valid"`
}

type FlowResponseNode struct {
	ID          uint                 `json:"id"`
	Title       string               `json:"title"`
	Type        string               `json:"type"`
	Thumbnail   string               `json:"thumbnail"`
	Description string               `json:"description"`
	CommentedAt time.Time            `json:"commented_at"`
	CreatedAt   time.Time            `json:"created_at"`
	Flow        models.NodeFlow      `json:"flow"`
	User        FlowNodeResponseUser `json:"user"`
}

func (r *FlowResponse) Init(
	before []models.Node,
	after []models.Node,
	heroes []models.Node,
	updated []models.Node,
	recent []models.Node,
	valid []uint,
) *FlowResponse {
	r.Before = make([]FlowResponseNode, len(before))
	r.After = make([]FlowResponseNode, len(after))
	r.Heroes = make([]FlowResponseNode, len(heroes))
	r.Updated = make([]FlowResponseNode, len(updated))
	r.Recent = make([]FlowResponseNode, len(recent))
	r.Valid = valid

	for k, v := range before {
		r.Before[k] = *new(FlowResponseNode).Init(v)
	}

	for k, v := range after {
		r.After[k] = *new(FlowResponseNode).Init(v)
	}

	for k, v := range heroes {
		r.Heroes[k] = *new(FlowResponseNode).Init(v)
	}

	for k, v := range updated {
		r.Updated[k] = *new(FlowResponseNode).Init(v)
	}

	for k, v := range recent {
		r.Recent[k] = *new(FlowResponseNode).Init(v)
	}

	return r
}

func (r *FlowResponseNode) Init(node models.Node) *FlowResponseNode {
	r.ID = node.ID
	r.Title = node.Title
	r.Type = node.Type
	r.Thumbnail = node.Thumbnail
	r.Description = node.Description
	r.CommentedAt = node.CommentedAt
	r.Flow = node.Flow
	r.CreatedAt = node.CreatedAt

	r.User.ID = node.User.ID
	r.User.Username = node.User.Username
	r.User.Photo = node.User.Photo.Url

	return r
}
