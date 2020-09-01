package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/controllers"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/utils"
)

type FlowRouter struct {
	nodeController controllers.NodeController
	api            utils.AppApi
}

func (fr *FlowRouter) Init(api utils.AppApi, db db.DB, config app.Config) {
	fr.api = api
	fr.nodeController = *new(controllers.NodeController).Init(db, config)
}

// FlowRouter for /node/*
func (fr *FlowRouter) Handle(r *gin.RouterGroup) {
	r.GET("/diff", fr.api.AuthOptional, fr.nodeController.GetDiff)
}
