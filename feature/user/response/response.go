package response

import (
	"github.com/muerwre/vault-golang/feature/notification/response"
	response2 "github.com/muerwre/vault-golang/feature/upload/response"
	"github.com/muerwre/vault-golang/feature/user/dto"
	"time"
)

type UserCheckCredentialsResponse struct {
	ID               uint                                   `json:"id"`
	Username         string                                 `json:"username"`
	Email            string                                 `json:"email"`
	Role             string                                 `json:"role"`
	Fullname         string                                 `json:"fullname"`
	Description      string                                 `json:"description"`
	LastSeen         time.Time                              `json:"last_seen"`
	LastSeenMessages time.Time                              `json:"last_seen_messages"`
	LastSeenBoris    time.Time                              `json:"last_seen_boris"`
	Photo            *response2.ShallowFileResponse         `json:"photo"`
	Cover            *response2.ShallowFileResponse         `json:"cover"`
	Notifications    *response.NotificationSettingsResponse `json:"notifications"`
}

func (uccr *UserCheckCredentialsResponse) FromDto(user *dto.UserDetailedDto) *UserCheckCredentialsResponse {
	uccr.ID = user.User.ID
	uccr.Username = user.User.Username
	uccr.Email = user.User.Email
	uccr.Role = user.User.Role
	uccr.Fullname = user.User.Fullname
	uccr.Description = user.User.Description
	uccr.LastSeen = user.User.LastSeen
	uccr.LastSeenMessages = user.User.LastSeenMessages
	uccr.LastSeenBoris = user.LastSeenBoris.Visited
	uccr.Notifications = new(response.NotificationSettingsResponse).FromModel(user.NotificationSettings)
	uccr.Cover = new(response2.ShallowFileResponse).FromModel(user.User.Cover)
	uccr.Photo = new(response2.ShallowFileResponse).FromModel(user.User.Photo)
	return uccr
}
