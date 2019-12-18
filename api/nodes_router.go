package api

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/controllers"
)

// UserRouter for /node/*
func NodesRouter(r *gin.RouterGroup, a *API) {
	r.GET("/diff", a.AuthOptional, controllers.Node.GetDiff)
}
