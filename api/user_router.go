package api

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/controllers"
)

// UserRouter for /user/*
func UserRouter(r *gin.RouterGroup, a *API) {
	controller := &controllers.UserController{
		Mailer: a.App.Mailer,
		DB:     a.App.DB,
		Config: a.App.Config,
	}

	r.POST("/login", controller.LoginUser)
	r.GET("/:username/profile", a.AuthOptional, controller.GetUserProfile)
	r.POST("/restore", controller.CreateRestoreCode)

	required := r.Group("/").Use(a.AuthRequired)
	{
		required.GET("/", a.WithUser(true), controller.CheckCredentials)
		required.PATCH("/", a.WithUser(false), controller.PatchUser)
	}
}
