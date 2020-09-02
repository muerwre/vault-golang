package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/controllers/usecase"
	"github.com/muerwre/vault-golang/db"
	"net/http"
)

type NotificationController struct {
	db      db.DB
	usecase usecase.NotificationUsecase
}

func (nc *NotificationController) Init(db db.DB) *NotificationController {
	nc.db = db
	nc.usecase = *new(usecase.NotificationUsecase).Init(db)

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
