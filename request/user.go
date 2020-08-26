package request

import (
	"reflect"

	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/passwords"
)

type UserPatchRequest struct {
	ID          uint
	Username    string `json:"username" validate:"omitempty,gte=3,lte=64"`
	Fullname    string `json:"fullname" validate:"omitempty,gte=1,lte=64"`
	Password    string `json:"password"`
	NewPassword string `json:"new_password" validate:"omitempty,gte=6,lte=64"`
	Email       string `json:"email" validate:"omitempty,email"`
	Description string `json:"description" validate:"omitempty,lte=512"`
	PhotoID     uint   `json:"-"`
	CoverID     uint   `json:"-"`

	Photo *models.File
	Cover *models.File
}

func (upd *UserPatchRequest) GetJsonTagName(f string) string {
	field, ok := reflect.TypeOf(upd).Elem().FieldByName(f)

	if !ok {
		return ""
	}

	return field.Tag.Get("json")
}

func (upd *UserPatchRequest) ApplyTo(u *models.User) {
	if upd.NewPassword != "" {
		u.Password, _ = passwords.HashPassword(upd.NewPassword)
	}

	u.Email = upd.Email
	u.Description = upd.Description
	u.Username = upd.Username
	u.Fullname = upd.Fullname

	if upd.PhotoID != 0 {
		u.PhotoID = &upd.PhotoID
	}
}

type UserCredentialsRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserRestoreCodeRequest struct {
	Field string `json:"field"`
}

type UserRestorePostRequest struct {
	Password string `json:"password"`
}

type UserMessage struct {
	Text string `json:"text"`
}

type UserMessageRequest struct {
	UserMessage `json:"message"`
}
