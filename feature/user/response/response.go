package response

import (
	"github.com/muerwre/vault-golang/db/models"
	"github.com/muerwre/vault-golang/feature/notification/response"
	"github.com/muerwre/vault-golang/feature/user/dto"
	"time"
)

type UserCheckCredentialsResponse struct {
	ID                   uint                                  `json:"id"`
	Username             string                                `json:"username"`
	Email                string                                `json:"email"`
	Role                 string                                `json:"role"`
	Fullname             string                                `json:"fullname"`
	Description          string                                `json:"description"`
	Cover                *models.File                          `json:"cover"`
	Photo                *models.File                          `json:"photo"`
	LastSeen             time.Time                             `json:"last_seen"`
	LastSeenMessages     time.Time                             `json:"last_seen_messages"`
	LastSeenBoris        time.Time                             `json:"last_seen_boris"`
	NotificationSettings response.NotificationSettingsResponse `json:"notificationSettings"`
}

func (uccr *UserCheckCredentialsResponse) FromDto(user *dto.UserDetailedDto) *UserCheckCredentialsResponse {
	uccr.ID = user.User.ID
	uccr.Username = user.User.Username
	uccr.Email = user.User.Email
	uccr.Role = user.User.Role
	uccr.Fullname = user.User.Fullname
	uccr.Description = user.User.Description
	uccr.Cover = user.User.Cover
	uccr.Photo = user.User.Photo
	uccr.LastSeen = user.User.LastSeen
	uccr.LastSeenMessages = user.User.LastSeenMessages
	uccr.LastSeenBoris = user.LastSeenBoris.Visited
	uccr.NotificationSettings = *new(response.NotificationSettingsResponse).FromModel(user.NotificationSettings)

	return uccr
}
