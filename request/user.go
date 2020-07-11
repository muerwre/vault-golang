package request

import (
	"github.com/muerwre/vault-golang/utils/validation"
	"reflect"

	"github.com/go-playground/validator"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
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

func (upd *UserPatchRequest) Validate(u *models.User, db db.DB) map[string]string {
	err := validation.On.Struct(upd)
	errors := map[string]string{}

	// We need password to change password or email or username
	if (upd.NewPassword != "" || (upd.Email != "" && upd.Email != u.Email) || (upd.Username != "" && upd.Username != u.Username)) &&
		(upd.Password == "" || !passwords.CheckPasswordHash(upd.Password, u.Password)) {
		errors[upd.GetJsonTagName("Password")] = codes.IncorrectPassword
	}

	// Shouldn't cover exist user
	if upd.Username != "" && upd.Username != u.Username &&
		db.First(&models.User{}, "username = ?", upd.Username).RowsAffected > 0 {
		errors[upd.GetJsonTagName("Username")] = codes.UserExist
	}

	// Photo should be at database
	if upd.Photo != nil && upd.Photo.ID != 0 {
		file := &models.File{}

		db.First(&file, "id = ?", upd.Photo.ID)

		if file == nil || file.UserID != u.ID || file.Type != models.FILE_TYPES.IMAGE {
			errors[upd.GetJsonTagName("Photo")] = codes.ImageConversionFailed
		}

		upd.PhotoID = file.ID
	}

	// Cover should be at database
	if upd.Cover != nil && upd.Cover.ID != 0 {
		file := &models.File{}

		db.First(&file, "id = ?", upd.Photo.ID)

		if file == nil || file.UserID != u.ID || file.Type != models.FILE_TYPES.IMAGE {
			errors[upd.GetJsonTagName("Cover")] = codes.ImageConversionFailed
		}

		upd.CoverID = file.ID
	}

	// Minimal requirements for fields from validate tag
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			field := upd.GetJsonTagName(err.Field())

			if field == "" {
				continue
			}

			if codes.VALIDATION_TO_CODE[err.Tag()] != "" {
				errors[field] = codes.VALIDATION_TO_CODE[err.Tag()]
			} else {
				errors[field] = codes.VALIDATION_TO_CODE["required"]
			}
		}
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
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
		u.PhotoID = upd.PhotoID
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
