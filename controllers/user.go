package controllers

import (
	"fmt"
	"net/http"
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

var User = &UserController{}

func (u *UserController) CheckCredentials(c *gin.Context) {
	user := c.MustGet("User").(*models.User)

	c.JSON(http.StatusOK, gin.H{"user": &user})
}

func (u *UserController) GetUserProfile(c *gin.Context) {
	username := c.Param("username")
	d := u.DB

	user, err := d.GetUserByUsername(username)

	if err != nil || user.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.USER_NOT_FOUND})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (u *UserController) LoginUser(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "" || password == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": codes.INCORRECT_DATA})
		return
	}

	d := c.MustGet("DB").(*db.DB)
	user, err := d.GetUserByUsername(username)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": codes.USER_NOT_FOUND})
		return
	}

	if !user.IsValidPassword(password) {
		md5hash := passwords.GetMD5Hash(password)

		if md5hash != user.Password {
			c.JSON(http.StatusUnauthorized, gin.H{"error": codes.USER_NOT_FOUND})
			return
		}
	}

	token := d.GenerateTokenFor(user)

	c.JSON(http.StatusOK, gin.H{"user": user, "token": token.Token})
}

func (uc *UserController) PatchUser(c *gin.Context) {
	d := c.MustGet("DB").(*db.DB)
	u := c.MustGet("User").(*models.User)

	data := &struct {
		User validation.UserPatchData `json:"user"`
	}{}

	err := c.ShouldBind(&data)

	if err != nil {
		fmt.Printf("ERR 1 %+v", err)
	}

	validation_errors := data.User.Validate(u, d)

	if validation_errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validation_errors})
		return
	}

	data.User.ApplyTo(u)

	d.Model(&models.User{}).Updates(u).Preload("Photo").Preload("Cover").First(&u)

	c.JSON(http.StatusOK, gin.H{"data": u})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.USER_NOT_FOUND})
		return
	}

	d.First(&user, "username = ? OR email = ?", params.Field, params.Field)

	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.USER_NOT_FOUND})
		return
	}

	code := &models.RestoreCode{
		UserID: user.ID,
		Code:   uuid.New().String(),
	}

	d.FirstOrCreate(&code, "UserId = ?", user.ID)

	if code.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.USER_NOT_FOUND})
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
		c.JSON(http.StatusNotFound, gin.H{"error": codes.CODE_IS_INVALID})
		return
	}

	code := &models.RestoreCode{}

	d.Preload("User").Preload("User.Photo").First(&code, "code = ?", id)

	if code.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.CODE_IS_INVALID})
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

	c.BindJSON(&params)

	if id == "" || params.Password == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.CODE_IS_INVALID})
		return
	}

	if len(params.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.TOO_SHIRT})
		return
	}

	code := &models.RestoreCode{}

	d.Preload("User").
		Preload("User.Photo").
		Preload("User.Cover").
		First(&code, "code = ?", id)

	if code.ID == 0 || code.UserID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.CODE_IS_INVALID})
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
		c.JSON(http.StatusNotFound, gin.H{"error": codes.USER_NOT_FOUND})
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
		c.JSON(http.StatusNotFound, gin.H{"error": codes.USER_NOT_FOUND})
		return
	}

	params := struct {
		Message string `json:"message"`
	}{}

	err = c.BindJSON(&params)

	if err != nil || user.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.INCORRECT_DATA})
		return
	}

	message := &models.Message{
		Text:   params.Message,
		FromID: u.ID,
		ToID:   user.ID,
	}

	if !message.IsValid() {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.INCORRECT_DATA})
		return
	}

	q := d.Model(&models.Message{}).Create(message)

	if q.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.INCORRECT_DATA, "details": q.Error.Error()})
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

	c.JSON(http.StatusOK, gin.H{"message": message})
}
