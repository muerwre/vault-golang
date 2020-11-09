package usecase

import (
	"fmt"
	"github.com/go-playground/validator"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/db/models"
	constants2 "github.com/muerwre/vault-golang/feature/node/constants"
	constants3 "github.com/muerwre/vault-golang/feature/upload/constants"
	"github.com/muerwre/vault-golang/feature/user/request"
	"github.com/muerwre/vault-golang/utils/codes"
	"github.com/muerwre/vault-golang/utils/passwords"
	"time"
)

type UserUsecase struct {
	db   db.DB
	user db.UserRepository
}

func (uc *UserUsecase) Init(db db.DB) *UserUsecase {
	uc.db = db
	uc.user = *db.User
	return uc
}

func (uc UserUsecase) ValidatePatchRequest(data *request.UserPatchRequest, u models.User) map[string]string {
	err := data.Validate()
	errors := map[string]string{}

	// We need password to change password or email or username
	if (data.NewPassword != "" || (data.Email != "" && data.Email != u.Email) || (data.Username != "" && data.Username != u.Username)) &&
		(data.Password == "" || !passwords.CheckPasswordHash(data.Password, u.Password)) {
		errors[data.GetJsonTagName("Password")] = codes.IncorrectPassword
	}

	// Shouldn't cover exist user
	if data.Username != "" && data.Username != u.Username {
		if _, err := uc.db.User.GetByUsername(data.Username); err == nil {
			errors[data.GetJsonTagName("Username")] = codes.UserExistWithUsername
		}
	}

	// Shouldn't cover exist user
	if data.Email != "" && data.Email != u.Email {
		if _, err := uc.db.User.GetByEmail(data.Email); err == nil {
			errors[data.GetJsonTagName("Email")] = codes.UserExistWithEmail
		}
	}

	// Photo should be at database
	if data.Photo != nil && data.Photo.ID != 0 {
		if file, err := uc.db.File.GetById(data.Photo.ID); err != nil || file.Type != constants3.FileTypeImage {
			errors[data.GetJsonTagName("Photo")] = codes.ImageConversionFailed
		} else {
			data.PhotoID = file.ID
		}
	}

	// Cover should be at database
	if data.Cover != nil && data.Cover.ID != 0 {
		if file, err := uc.db.File.GetById(data.Photo.ID); err != nil || file.Type != constants3.FileTypeImage {
			errors[data.GetJsonTagName("Cover")] = codes.ImageConversionFailed
		} else {
			data.CoverID = file.ID
		}
	}

	// Minimal requirements for fields from validate tag
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			field := data.GetJsonTagName(err.Field())

			if field == "" {
				continue
			}

			if codes.ValidationToCode[err.Tag()] != "" {
				errors[field] = codes.ValidationToCode[err.Tag()]
			} else {
				errors[field] = codes.ValidationToCode["required"]
			}
		}
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}

func (uc UserUsecase) GetUserForCheckCredentials(uid uint) (user *models.User, lastSeenBoris *time.Time, err error) {
	user, err = uc.db.User.GetById(uid)

	if err != nil {
		return nil, nil, err
	}

	view, err := uc.db.NodeView.GetOrCreateOne(uid, constants2.BorisNodeId)

	if err != nil {
		return nil, nil, err
	}

	return user, &view.Visited, nil
}

func (uc UserUsecase) FillMessageFromData(from models.User, recp string, data request.UserMessageRequest) (*models.Message, error) {
	to, err := uc.db.User.GetByUsername(recp)

	if err != nil {
		return nil, fmt.Errorf(codes.UserNotFound)
	}

	message := &models.Message{}

	if data.Message.ID != 0 {
		message, err = uc.db.Message.LoadMessageWithUsers(data.Message.ID)

		if err != nil {
			return nil, err
		}

		if message.From.ID != from.ID || message.To.ID != to.ID {
			return nil, fmt.Errorf(codes.NotEnoughRights)
		}
	} else {
		message.FromID = &from.ID
		message.ToID = &to.ID
	}

	message.Text = data.Message.Text

	if !message.IsValid() {
		return nil, fmt.Errorf(codes.IncorrectData)
	}

	return message, nil
}

func (uc UserUsecase) SaveMessage(message *models.Message) error {
	if message.ID == 0 {
		if err := uc.db.Message.CreateMessage(message); err != nil {
			return err
		}
	} else {
		if err := uc.db.Message.SaveMessage(message); err != nil {
			return err
		}
	}

	if m, err := uc.db.Message.LoadMessageWithUsers(message.ID); err != nil {
		return err
	} else {
		*message = *m
	}

	return nil
}

func (uc UserUsecase) UpdateMessageView(fromID uint, toID uint) error {
	view := &models.MessageView{
		DialogId: toID,
		UserId:   fromID,
	}

	if err := uc.db.Where("userId = ? AND dialogId = ?", fromID, toID).FirstOrCreate(&view).Error; err != nil {
		return err
	}

	view.Viewed = time.Now()

	return uc.db.Save(&view).Error
}

func (uc UserUsecase) GetMessagesForUsers(fromID uint, toID uint, after time.Time, before time.Time, limit int) ([]models.Message, error) {
	messages := []models.Message{}

	err := uc.db.Preload("From").
		Preload("To").
		Where("(fromId = ? AND toId = ?) OR (fromId = ? AND toId = ?)", fromID, toID, toID, fromID).
		Where("created_at > ? AND created_at < ?", after, before).
		Limit(limit).
		Order("created_at DESC").
		Find(&messages).Error

	return messages, err
}

func (uc UserUsecase) GetByEmail(email string) (*models.User, error) {
	return uc.user.GetByEmail(email)
}

func (uc UserUsecase) GetByUsername(username string) (*models.User, error) {
	return uc.user.GetByUsername(username)
}

func (uc UserUsecase) GenerateTokenFor(u *models.User) (*models.Token, error) {
	return uc.user.GenerateTokenFor(u)
}

func (uc UserUsecase) CreateUser(user *models.User) error {
	return uc.user.Create(user)
}

func (uc UserUsecase) UpdateUserPhoto(user *models.User, photo *models.File) error {
	return uc.user.UpdatePhoto(user.ID, photo.ID)
}
