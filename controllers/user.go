package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
	"github.com/muerwre/vault-golang/utils/validation"
)

type UserController struct{}

var User = &UserController{}

func (u *UserController) CheckCredentials(c *gin.Context) {
	user := c.MustGet("User").(*models.User)

	c.JSON(http.StatusOK, gin.H{"user": &user})
}

func (u *UserController) GetUserProfile(c *gin.Context) {
	username := c.Param("username")
	d := c.MustGet("DB").(*db.DB)

	user, err := d.GetUserByUsername(username)

	if err != nil {
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": codes.USER_NOT_FOUND})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
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

	validation_errors := data.User.Validate(*u, d)

	if validation_errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validation_errors})
		return
	}

	d.Model(&u).Save(u)

	c.JSON(http.StatusOK, gin.H{"data": u})
}
