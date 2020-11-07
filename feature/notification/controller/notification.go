package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/db"
	usecase2 "github.com/muerwre/vault-golang/feature/notification/usecase"
	"net/http"
)

type NotificationController struct {
	db      db.DB
	usecase usecase2.NotificationUsecase
}

func (nc *NotificationController) Init(db db.DB) *NotificationController {
	nc.db = db
	nc.usecase = *new(usecase2.NotificationUsecase).Init(db)

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
