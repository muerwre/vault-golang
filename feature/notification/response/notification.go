package response

import (
	"github.com/muerwre/vault-golang/models"
	"time"
)

var NotificationTypes = struct {
	Message string
}{
	Message: "message",
}

type Content interface {
}

type Notification struct {
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	Content   `json:"content"`
}

func (n *Notification) FromMessage(m models.Message) {
	n.Type = NotificationTypes.Message
	n.CreatedAt = m.CreatedAt
	n.Content = m
}
