package usecase

import (
	"fmt"
	"github.com/go-playground/validator"
	"github.com/muerwre/vault-golang/constants"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/request"
	"github.com/muerwre/vault-golang/utils/codes"
	"github.com/muerwre/vault-golang/utils/passwords"
	"github.com/muerwre/vault-golang/utils/validation"
	"time"
)

type UserUsecase struct {
	db db.DB
}

func (uc *UserUsecase) Init(db db.DB) *UserUsecase {
	uc.db = db
	return uc
}

func (uc UserUsecase) ValidatePatchRequest(data *request.UserPatchRequest, u models.User) map[string]string {
	err := validation.On.Struct(data)
	errors := map[string]string{}

	// We need password to change password or email or username
	if (data.NewPassword != "" || (data.Email != "" && data.Email != u.Email) || (data.Username != "" && data.Username != u.Username)) &&
		(data.Password == "" || !passwords.CheckPasswordHash(data.Password, u.Password)) {
		errors[data.GetJsonTagName("Password")] = codes.IncorrectPassword
	}

	// Shouldn't cover exist user
	if data.Username != "" && data.Username != u.Username {
		if _, err := uc.db.UserRepository.GetByUsername(data.Username); err == nil {
			errors[data.GetJsonTagName("Username")] = codes.UserExistWithUsername
		}
	}

	// Shouldn't cover exist user
	if data.Email != "" && data.Email != u.Email {
		if _, err := uc.db.UserRepository.GetByEmail(data.Email); err == nil {
			errors[data.GetJsonTagName("Email")] = codes.UserExistWithEmail
		}
	}

	// Photo should be at database
	if data.Photo != nil && data.Photo.ID != 0 {
		if file, err := uc.db.FileRepository.GetById(data.Photo.ID); err != nil || file.Type != constants.FileTypeImage {
			errors[data.GetJsonTagName("Photo")] = codes.ImageConversionFailed
		} else {
			data.PhotoID = file.ID
		}
	}

	// Cover should be at database
	if data.Cover != nil && data.Cover.ID != 0 {
		if file, err := uc.db.FileRepository.GetById(data.Photo.ID); err != nil || file.Type != constants.FileTypeImage {
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
	user, err = uc.db.UserRepository.GetById(uid)

	if err != nil {
		return nil, nil, err
	}

	view, err := uc.db.NodeViewRepository.GetOne(uid, constants.BorisNodeId)

	if err != nil {
		return nil, nil, err
	}

	return user, &view.Visited, nil
}

func (uc UserUsecase) FillMessageFromData(from models.User, recp string, data request.UserMessageRequest) (*models.Message, error) {
	to, err := uc.db.UserRepository.GetByUsername(recp)

	if err != nil {
		return nil, fmt.Errorf(codes.UserNotFound)
	}

	message := &models.Message{}

	if data.ID != 0 {
		message, err = uc.db.MessageRepository.LoadMessageWithUsers(data.ID)

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

	message.Text = data.UserMessage.Text

	if !message.IsValid() {
		return nil, fmt.Errorf(codes.IncorrectData)
	}

	return message, nil
}

func (uc UserUsecase) SaveMessage(message *models.Message) error {
	if message.ID == 0 {
		if err := uc.db.MessageRepository.CreateMessage(message); err != nil {
			return err
		}
	} else {
		if err := uc.db.MessageRepository.SaveMessage(message); err != nil {
			return err
		}
	}

	if m, err := uc.db.MessageRepository.LoadMessageWithUsers(message.ID); err != nil {
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

func (uc UserUsecase) GetMessagesForUsers(fromID uint, toID uint) ([]models.Message, error) {
	messages := []models.Message{}

	err := uc.db.Preload("From").
		Preload("To").
		Where("(fromId = ? AND toId = ?) OR (fromId = ? AND toId = ?)", fromID, toID, toID, fromID).
		Limit(50).
		Order("created_at DESC").
		Find(&messages).Error

	return messages, err
}
