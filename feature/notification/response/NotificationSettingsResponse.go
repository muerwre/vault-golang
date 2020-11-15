package response

import (
	"github.com/muerwre/vault-golang/db/models"
	"time"
)

type NotificationSettingsResponse struct {
	SubscribedToFlow      bool      `json:"subscribed_to_flow"`
	SubscribedToComments  bool      `json:"subscribed_to_comments"`
	LastSeenNotifications time.Time `json:"last_seen_notifications"`
}

func (r *NotificationSettingsResponse) FromModel(m *models.NotificationSettings) *NotificationSettingsResponse {
	r.SubscribedToComments = m.SubscribedToComments
	r.SubscribedToFlow = m.SubscribedToFlow
	r.LastSeenNotifications = m.LastSeenNotifications
	return r
}
