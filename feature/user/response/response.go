package response

import (
	"github.com/muerwre/vault-golang/models"
	"time"
)

type UserCheckCredentialsResponse struct {
	ID               uint         `json:"id"`
	Username         string       `json:"username"`
	Email            string       `json:"email"`
	Role             string       `json:"role"`
	Fullname         string       `json:"fullname"`
	Description      string       `json:"description"`
	Cover            *models.File `json:"cover"`
	Photo            *models.File `json:"photo"`
	LastSeen         time.Time    `json:"last_seen"`
	LastSeenMessages time.Time    `json:"last_seen_messages"`
	LastSeenBoris    time.Time    `json:"last_seen_boris"`
}

func (uccr *UserCheckCredentialsResponse) Init(user *models.User, lastSeenBoris time.Time) *UserCheckCredentialsResponse {
	uccr.ID = user.ID
	uccr.Username = user.Username
	uccr.Email = user.Email
	uccr.Role = user.Role
	uccr.Fullname = user.Fullname
	uccr.Description = user.Description
	uccr.Cover = user.Cover
	uccr.Photo = user.Photo
	uccr.LastSeen = user.LastSeen
	uccr.LastSeenMessages = user.LastSeenMessages
	uccr.LastSeenBoris = lastSeenBoris

	return uccr
}
