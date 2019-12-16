package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
)

type UserController struct{}

var User = &UserController{}

func (u *UserController) CheckCredentials(c *gin.Context) {
	user := c.MustGet("User").(*models.User)

	c.JSON(http.StatusOK, gin.H{"user": &user})
	return
}

func (u *UserController) GetUserProfile(c *gin.Context) {
	username := c.Param("username")
	d := c.MustGet("DB").(*db.DB)

	user, err := d.GetUserByUsername(username)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.USER_NOT_FOUND})
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}
