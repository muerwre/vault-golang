package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/feature/node/controller"
	"github.com/muerwre/vault-golang/service/notification"
	"github.com/muerwre/vault-golang/utils"
)

type FlowRouter struct {
	nodeController controller.NodeController
	api            utils.AppApi
}

func (fr *FlowRouter) Init(api utils.AppApi, db db.DB, config app.Config, notifier notification.NotificationService) {
	fr.api = api
	fr.nodeController = *new(controller.NodeController).Init(db, config, notifier)
}

// FlowRouter for /node/*
func (fr *FlowRouter) Handle(r *gin.RouterGroup) {
	r.GET("/diff", fr.api.AuthOptional, fr.nodeController.GetDiff)
}
