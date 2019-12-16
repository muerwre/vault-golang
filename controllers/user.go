package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/models"
)

type UserController struct{}

var User = &UserController{}

func (u *UserController) CheckCredentials(c *gin.Context) {
	user := c.MustGet("User").(*models.User)

	c.JSON(http.StatusOK, gin.H{"user": &user})
	return
}
