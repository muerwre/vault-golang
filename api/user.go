package api

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/controllers"
)

// UserRouter for /user/*
func UserRouter(r *gin.RouterGroup, a *API) {
	authorized := r.Group("/").Use(a.AuthRequired)
	{
		authorized.GET("/", controllers.User.CheckCredentials)
	}
}
