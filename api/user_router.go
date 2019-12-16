package api

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/controllers"
)

// UserRouter for /user/*
func UserRouter(r *gin.RouterGroup, a *API) {
	r.GET("/:username/profile", controllers.User.GetUserProfile)

	authorized := r.Group("/").Use(a.AuthRequired)
	{
		authorized.GET("/", controllers.User.CheckCredentials)
	}
}
