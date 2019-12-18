package api

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/controllers"
)

// UserRouter for /user/*
func UserRouter(r *gin.RouterGroup, a *API) {
	r.POST("/login", controllers.User.LoginUser)
	r.GET("/:username/profile", a.AuthOptional, controllers.User.GetUserProfile)

	authorized := r.Group("/").Use(a.AuthRequired)
	{
		authorized.GET("/", a.WithUser(true), controllers.User.CheckCredentials)
		authorized.PATCH("/", a.WithUser(false), controllers.User.PatchUser)
	}
}
