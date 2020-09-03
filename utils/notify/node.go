package notify

import (
	"github.com/muerwre/vault-golang/models"
	"time"
)

type NodeNotification struct {
	Type      string       `json:"type"`
	Node      *models.Node `json:"node"`
	CreatedAt time.Time    `json:"created_at"`
}

func (n NodeNotification) GetType() string {
	return n.Type
}

func (n NodeNotification) GetContent() interface{} {
	return n.Node
}

func (n NodeNotification) GetCreatedAt() time.Time {
	return n.CreatedAt
}
