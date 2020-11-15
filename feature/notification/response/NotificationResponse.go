package response

import (
	"github.com/muerwre/vault-golang/db/models"
	"time"
)

const (
	NotificationTypeMessage = "message"
)

type Content interface {
}

type NotificationResponse struct {
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	Content   `json:"content"`
}

func (n *NotificationResponse) FromMessage(m models.Message) {
	n.Type = NotificationTypeMessage
	n.CreatedAt = m.CreatedAt
	n.Content = m
}
