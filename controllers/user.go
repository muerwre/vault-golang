package controllers

import (
	"github.com/muerwre/vault-golang/response"
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
	"github.com/muerwre/vault-golang/utils/validation"
)

type UserController struct {
	Mailer *mail.Mailer
	DB     *db.DB
	Config *app.Config
}

func (uc *UserController) CheckCredentials(c *gin.Context) {
	user := c.MustGet("User").(*models.User)

	c.JSON(http.StatusOK, gin.H{"user": &user})
}

func (uc *UserController) GetUserProfile(c *gin.Context) {
	username := c.Param("username")
	d := uc.DB

	user, err := d.GetUserByUsername(username)

	if err != nil || user.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.UserNotFound})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (uc *UserController) LoginUser(c *gin.Context) {
	credentials := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}

	err := c.BindJSON(&credentials)

	if err != nil || credentials.Username == "" || credentials.Password == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": codes.IncorrectData})
		return
	}

	d := uc.DB
	user, err := d.GetUserByUsername(credentials.Username)

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

	token := d.GenerateTokenFor(user)

	c.JSON(http.StatusOK, gin.H{"user": user, "token": token.Token})
}

func (uc *UserController) PatchUser(c *gin.Context) {
	d := c.MustGet("DB").(*db.DB)
	u := c.MustGet("User").(*models.User)

	data := struct {
		User validation.UserPatchData `json:"user"`
	}{}

	err := c.ShouldBind(&data)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
	}

	validationErrors := data.User.Validate(u, d)

	if validationErrors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
		return
	}

	data.User.ApplyTo(u)

	d.Model(&models.User{}).Updates(u).Preload("Photo").Preload("Cover").First(&u)

	c.JSON(http.StatusOK, gin.H{"user": u})
}

func (uc *UserController) CreateRestoreCode(c *gin.Context) {
	user := &models.User{}
	d := uc.DB
	mailer := uc.Mailer
	config := uc.Config

	params := struct {
		Field string `json:"field"`
	}{}

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

	params := struct {
		Password string `json:"password"`
	}{}

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

	d.Set("gorm:association_autoupdate", false).
		Set("gorm:association_save_reference", false).
		Model(&models.User{}).Where("id = ?", code.UserID).
		Update("password", params.Password)

	d.Delete(&code)

	token := d.GenerateTokenFor(code.User)

	c.JSON(http.StatusOK, gin.H{"user": code.User, "token": token.Token})
}

func (uc *UserController) GetUserMessages(c *gin.Context) {
	username := c.Param("username")
	u := c.MustGet("User").(*models.User)
	d := uc.DB

	user, err := d.GetUserByUsername(username)

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

	user, err := d.GetUserByUsername(username)

	if err != nil || user.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.UserNotFound})
		return
	}

	type Message struct {
		Text string `json:"text"`
	}

	params := struct {
		Message `json:"message"`
	}{}

	err = c.BindJSON(&params)

	if err != nil || user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	message := &models.Message{
		Text:   params.Message.Text,
		FromID: u.ID,
		ToID:   user.ID,
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

	messages, err := d.GetUserNewMessages(*user, exclude, last)

	boris, _ := d.GetNodeBoris()
	notifications := make([]response.Notification, len(messages))

	for k := range notifications {
		notifications[k].FromMessage(messages[k])
	}

	c.JSON(http.StatusOK, gin.H{"notifications": notifications, "boris": boris})
}
