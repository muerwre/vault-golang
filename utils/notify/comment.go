package notify

import (
	"github.com/muerwre/vault-golang/models"
	"time"
)

type CommentNotification struct {
	Type      string          `json:"type"`
	Comment   *models.Comment `json:"comment"`
	CreatedAt time.Time       `json:"created_at"`
}

func (n CommentNotification) GetType() string {
	return n.Type
}

func (n CommentNotification) GetContent() interface{} {
	return n.Comment
}

func (n CommentNotification) GetCreatedAt() time.Time {
	return n.CreatedAt
}
