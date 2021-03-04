package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/internal/db"
	"github.com/muerwre/vault-golang/internal/feature/notification/request"
	"github.com/muerwre/vault-golang/internal/feature/notification/response"
	"github.com/muerwre/vault-golang/internal/feature/notification/usecase"
	"github.com/muerwre/vault-golang/pkg/codes"
	"github.com/sirupsen/logrus"
	"net/http"
)

type NotificationController struct {
	notification usecase.NotificationUsecase
}

func (nc *NotificationController) Init(db db.DB) *NotificationController {
	nc.notification = *new(usecase.NotificationUsecase).Init(db)
	return nc
}

// GetNotifications returns user notifications
func (nc NotificationController) GetNotifications(c *gin.Context) {
	u := c.MustGet("UID").(uint)

	settings, err := nc.notification.GetUserNotificationSettings(u)
	if err != nil {
		logrus.Warnf("Can't get user notifications settings for %s: %s", u, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.CantGetNotificationSettings})
		return
	}

	res := new(response.NotificationSettingsResponse).FromModel(settings)

	c.JSON(http.StatusOK, res)
}

// PostSettings changes user subscription settings
func (nc NotificationController) PostSettings(c *gin.Context) {
	uid := c.MustGet("UID").(uint)

	req := &request.NotificationSettingsRequest{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	settings, err := nc.notification.UpdateUserNotificationSettings(uid, req)
	if err != nil {
		logrus.Warnf("Can't update user notifications settings for %s: %s", uid, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.CantUpdateNotificationSettings})
		return
	}

	res := new(response.NotificationSettingsResponse).FromModel(settings)

	c.JSON(http.StatusOK, res)
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
