package controllers

import (
	"github.com/muerwre/vault-golang/controllers/usecase"
	"github.com/muerwre/vault-golang/request"
	"github.com/muerwre/vault-golang/response"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
	"github.com/muerwre/vault-golang/utils/mail"
	"github.com/muerwre/vault-golang/utils/passwords"
)

type UserController struct {
	Mailer mail.Mailer
	DB     db.DB
	Config app.Config

	usecase usecase.UserUsecase
}

func (uc *UserController) Init(db db.DB, mailer mail.Mailer, config app.Config) *UserController {
	uc.DB = db
	uc.Mailer = mailer
	uc.Config = config
	uc.usecase = *new(usecase.UserUsecase).Init(db)

	return uc
}

func (uc *UserController) CheckCredentials(c *gin.Context) {
	user := c.MustGet("User").(*models.User)

	user, lastSeenBoris, err := uc.usecase.GetUserForCheckCredentials(user.ID)

	if err != nil {
		logrus.Warnf("Can't load current user %s:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.CantLoadUser})
		return
	}

	resp := new(response.UserCheckCredentialsResponse).Init(user, *lastSeenBoris)

	uc.DB.UserRepository.UpdateLastSeen(user)

	c.JSON(http.StatusOK, gin.H{"user": &resp})
}

func (uc *UserController) GetUserProfile(c *gin.Context) {
	username := c.Param("username")
	d := uc.DB

	user, err := d.UserRepository.GetByUsername(username)

	if err != nil || user.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.UserNotFound})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (uc *UserController) LoginUser(c *gin.Context) {
	credentials := request.UserCredentialsRequest{}

	err := c.BindJSON(&credentials)

	if err != nil || credentials.Username == "" || credentials.Password == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": codes.IncorrectData})
		return
	}

	d := uc.DB
	user, err := d.UserRepository.GetByUsername(credentials.Username)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": codes.UserNotFound})
		return
	}

	if !user.IsValidPassword(credentials.Password) {
		md5hash := passwords.GetMD5Hash(credentials.Password)

		if md5hash != user.Password {
			c.JSON(http.StatusUnauthorized, gin.H{"error": codes.UserNotFound})
			return
		}
	}

	user, lastSeenBoris, err := uc.usecase.GetUserForCheckCredentials(user.ID)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": codes.UserNotFound})
		return
	}

	resp := new(response.UserCheckCredentialsResponse).Init(user, *lastSeenBoris)
	token := d.UserRepository.GenerateTokenFor(user)

	c.JSON(http.StatusOK, gin.H{"user": resp, "token": token.Token})
}

func (uc *UserController) PatchUser(c *gin.Context) {
	u := c.MustGet("User").(*models.User)

	data := &request.UserPatchRequest{}

	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	errors := uc.usecase.ValidatePatchRequest(data, *u)

	if errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	data.ApplyTo(u)

	if err := uc.DB.UserRepository.Save(u); err != nil {
		logrus.Infof("Can't update user: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.CantSaveUser})
		return
	}

	uc.CheckCredentials(c)

	return
}

func (uc *UserController) CreateRestoreCode(c *gin.Context) {
	user := &models.User{}
	d := uc.DB
	mailer := uc.Mailer
	config := uc.Config

	params := request.UserRestoreCodeRequest{}

	err := c.BindJSON(&params)

	if err != nil || params.Field == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.UserNotFound})
		return
	}

	d.First(&user, "username = ? OR email = ?", params.Field, params.Field)

	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.UserNotFound})
		return
	}

	code := &models.RestoreCode{
		UserID: user.ID,
		Code:   uuid.New().String(),
	}

	d.FirstOrCreate(&code, "UserId = ?", user.ID)

	if code.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.UserNotFound})
		return
	}

	message := mailer.Create(
		user.Email,
		mail.MAIL_RESTORE_SUBJECT,
		mail.MAIL_RESTORE_TEXT,
		mail.MAIL_RESTORE_HTML,
		&map[string]string{
			"url":  config.Protocol + "://" + config.PublicHost + config.ResetUrl,
			"code": code.Code,
		},
	)

	mailer.Chan <- message

	c.JSON(http.StatusCreated, gin.H{})
}

func (uc UserController) GetRestoreCode(c *gin.Context) {
	id := c.Param("id")
	d := uc.DB

	if id == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.CodeIsInvalid})
		return
	}

	code := &models.RestoreCode{}

	d.Preload("User").Preload("User.Photo").First(&code, "code = ?", id)

	if code.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.CodeIsInvalid})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user": map[string]interface{}{
			"username": code.User.Username,
			"photo":    code.User.Photo,
		},
	})
}

func (uc UserController) PostRestoreCode(c *gin.Context) {
	id := c.Param("id")
	d := uc.DB

	params := request.UserRestorePostRequest{}

	err := c.BindJSON(&params)

	if err != nil || id == "" || params.Password == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.CodeIsInvalid})
		return
	}

	if len(params.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.TooShirt})
		return
	}

	code := &models.RestoreCode{}

	d.Preload("User").
		Preload("User.Photo").
		Preload("User.Cover").
		First(&code, "code = ?", id)

	if code.ID == 0 || code.UserID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.CodeIsInvalid})
		return
	}

	password, _ := passwords.HashPassword(params.Password)

	d.Set("gorm:association_autoupdate", false).
		Set("gorm:association_save_reference", false).
		Model(&models.User{}).
		Where("id = ?", code.UserID).
		Update("password", password)

	d.Delete(&code, "id = ?", code.ID)

	token := d.UserRepository.GenerateTokenFor(code.User)

	c.JSON(http.StatusOK, gin.H{"user": code.User, "token": token.Token})
}

func (uc *UserController) GetUserMessages(c *gin.Context) {
	username := c.Param("username")
	u := c.MustGet("User").(*models.User)
	d := uc.DB

	user, err := d.UserRepository.GetByUsername(username)

	if err != nil || user.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.UserNotFound})
		return
	}

	messages := []models.Message{}

	d.Preload("From").
		Preload("To").
		Where("(fromId = ? AND toId = ?) OR (fromId = ? AND toId = ?)", u.ID, user.ID, user.ID, u.ID).
		Limit(50).
		Order("created_at DESC").
		Find(&messages)

	view := &models.MessageView{
		DialogId: user.ID,
		UserId:   u.ID,
	}

	d.
		Where("userId = ? AND dialogId = ?", u.ID, user.ID).
		FirstOrCreate(&view)

	view.Viewed = time.Now()

	d.Save(&view)

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (uc *UserController) PostMessage(c *gin.Context) {
	username := c.Param("username")
	u := c.MustGet("User").(*models.User)
	d := uc.DB

	user, err := d.UserRepository.GetByUsername(username)

	if err != nil || user.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.UserNotFound})
		return
	}

	params := request.UserMessageRequest{}

	err = c.BindJSON(&params)

	if err != nil || user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	message := &models.Message{
		Text:   params.UserMessage.Text,
		FromID: &u.ID,
		ToID:   &user.ID,
	}

	if !message.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.TooShirt})
		return
	}

	q := d.Model(&models.Message{}).Create(message)

	if q.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.IncorrectData, "details": q.Error.Error()})
		return
	}

	view := &models.MessageView{
		DialogId: user.ID,
		UserId:   u.ID,
	}

	d.
		Where("userId = ? AND dialogId = ?", u.ID, user.ID).
		FirstOrCreate(&view)

	view.Viewed = time.Now()

	d.Save(&view)

	d.Where("id = ?", message.ID).Preload("From").Preload("To").First(&message)

	c.JSON(http.StatusOK, gin.H{"message": message})
}

func (uc *UserController) GetUpdates(c *gin.Context) {
	d := uc.DB
	user := c.MustGet("User").(*models.User)
	last := c.Query("last")
	exclude, err := strconv.Atoi(c.Query("exclude_dialogs"))

	if err != nil {
		exclude = 0
	}

	messages, err := d.UserRepository.GetUserNewMessages(*user, exclude, last)

	boris, _ := d.NodeRepository.GetNodeBoris()
	notifications := make([]response.Notification, len(messages))

	for k := range notifications {
		notifications[k].FromMessage(messages[k])
	}

	c.JSON(http.StatusOK, gin.H{"notifications": notifications, "boris": boris})
}
