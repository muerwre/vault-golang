package request

import (
	"github.com/go-playground/validator"
	"github.com/muerwre/vault-golang/internal/db/models"
	"github.com/muerwre/vault-golang/pkg/passwords"
	"reflect"
	"time"
)

type UserPatchRequest struct {
	ID               uint
	Username         string     `json:"username" validate:"omitempty,gte=3,lte=64"`
	Fullname         string     `json:"fullname" validate:"omitempty,gte=1,lte=64"`
	Password         string     `json:"password"`
	NewPassword      string     `json:"new_password" validate:"omitempty,gte=6,lte=64"`
	Email            string     `json:"email" validate:"omitempty,gte=2,email"`
	Description      string     `json:"description" validate:"omitempty,lte=512"`
	PhotoID          uint       `json:"-"`
	CoverID          uint       `json:"-"`
	LastSeenMessages *time.Time `json:"last_seen_messages"`

	Photo *models.File
	Cover *models.File
}

func (upd *UserPatchRequest) Validate() error {
	return validator.New().Struct(upd)
}

func (upd *UserPatchRequest) GetJsonTagName(f string) string {
	field, ok := reflect.TypeOf(upd).Elem().FieldByName(f)

	if !ok {
		return ""
	}

	return field.Tag.Get("json")
}

func (upd *UserPatchRequest) ApplyTo(u *models.User) {
	u.Description = upd.Description
	u.Fullname = upd.Fullname

	if upd.NewPassword != "" {
		u.Password, _ = passwords.HashPassword(upd.NewPassword)
	}

	if upd.Email != "" {
		u.Email = upd.Email
	}

	if upd.Username != "" {
		u.Username = upd.Username
	}

	if upd.PhotoID != 0 {
		u.PhotoID = &upd.PhotoID
	}

	if upd.LastSeenMessages != nil {
		u.LastSeenMessages = *upd.LastSeenMessages
	}
}
