package dto

import "github.com/muerwre/vault-golang/internal/db/models"

type UserDetailedDto struct {
	User                 *models.User
	LastSeenBoris        *models.NodeView
	NotificationSettings *models.NotificationSettings
}
