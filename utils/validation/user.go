package validation

import (
	"reflect"

	"github.com/go-playground/validator"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
	"github.com/muerwre/vault-golang/utils/passwords"
)

type UserPatchData struct {
	Username    string `json:"username" validate:"omitempty,gte=3,lte=64"`
	Fullname    string `json:"fullname" validate:"omitempty,gte=1,lte=64"`
	Password    string `json:"password"`
	NewPassword string `json:"new_password" validate:"omitempty,gte=6,lte=64"`
	Email       string `json:"email" validate:"omitempty,email"`
	Description string `json:"description" validate:"omitempty,lte=512"`
	Photo       *models.File
}

func (d *UserPatchData) GetJsonTagName(f string) string {
	field, ok := reflect.TypeOf(d).Elem().FieldByName(f)

	if !ok {
		return ""
	}

	return field.Tag.Get("json")
}

func (d *UserPatchData) Validate(u *models.User, db *db.DB) map[string]string {
	err := On.Struct(d)
	errors := map[string]string{}

	// We need password to change password or email
	if (d.NewPassword != "" || (d.Email != "" && d.Email != u.Email)) &&
		(d.Password == "" || !passwords.CheckPasswordHash(d.Password, u.Password)) {
		errors[d.GetJsonTagName("Password")] = codes.INCORRECT_PASSWORD
	}

	// Shouldn't cover exist user
	if d.Username != "" && d.Username != u.Username &&
		db.First(&models.User{}, "username = ?", d.Username).RowsAffected > 0 {
		errors[d.GetJsonTagName("Username")] = codes.USER_EXIST
	}

	// Photo should be at database
	if d.Photo.ID != 0 {
		file := &models.File{}

		db.First(&file, "id = ?", d.Photo.ID)

		if file == nil || file.UserID != u.ID || file.Type != models.FILE_TYPES["IMAGE"] {
			errors[d.GetJsonTagName("Photo")] = codes.IMAGE_CONVERSION_FAILED
		}

		// d.Photo = &models.File{}
		// d.Photo = file
		// d.PhotoID = file.ID
	}

	// Minimal requirements for fields from validate tag
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			field := d.GetJsonTagName(err.Field())

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
