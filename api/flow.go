package api

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/controllers"
)

// FlowRouter for /node/*
func FlowRouter(r *gin.RouterGroup, a *API) {
	// TODO: replace with FlowController
	controller := &controllers.NodeController{
		DB: a.App.DB,
	}

	r.GET("/diff", a.AuthOptional, controller.GetDiff)
}
