package request

import (
	"github.com/muerwre/vault-golang/internal/db/models"
)

type NotificationSettingsRequest struct {
	SubscribedToFlow     bool `json:"subscribed_to_flow"`
	SubscribedToComments bool `json:"subscribed_to_comments"`
}

func (r NotificationSettingsRequest) ToModel() *models.NotificationSettings {
	return &models.NotificationSettings{
		SubscribedToComments: r.SubscribedToComments,
		SubscribedToFlow:     r.SubscribedToFlow,
	}
}
