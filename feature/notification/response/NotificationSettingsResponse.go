package response

import (
	"github.com/muerwre/vault-golang/db/models"
	"time"
)

type NotificationSettingsResponse struct {
	SubscribedToFlow      bool      `json:"flow"`
	SubscribedToComments  bool      `json:"comments"`
	LastSeenNotifications time.Time `json:"last_seen"`
}

func (r *NotificationSettingsResponse) FromModel(m *models.NotificationSettings) *NotificationSettingsResponse {
	if m == nil {
		return nil
	}

	r.SubscribedToComments = m.SubscribedToComments
	r.SubscribedToFlow = m.SubscribedToFlow
	r.LastSeenNotifications = m.LastSeenNotifications
	return r
}
