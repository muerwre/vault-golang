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

func (nc NotificationController) NodeGet(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func (nc NotificationController) NodeWatch(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func (nc NotificationController) NodeUnwatch(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}
