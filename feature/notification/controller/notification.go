package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/feature/notification/usecase"
	"net/http"
)

type NotificationController struct {
	notification usecase.NotificationUsecase
}

func (nc *NotificationController) Init(db db.DB) *NotificationController {
	nc.notification = *new(usecase.NotificationUsecase).Init(db)
	return nc
}

// PostSettings changes user subscription settings
func (nc NotificationController) PostSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

// NodeGet returns node notification subscription status
func (nc NotificationController) NodeGet(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

// NodePost creates node notification subscription
func (nc NotificationController) NodePost(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

// NodeDelete removes node notification subscription
func (nc NotificationController) NodeDelete(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}
