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

	r.GET("/restore/:id", controller.GetRestoreCode)
	r.POST("/restore/:id", controller.PostRestoreCode)
	r.POST("/restore", controller.CreateRestoreCode)

	optional := r.Group("/").Use(a.AuthOptional)
	{
		optional.GET("/user/:username/profile", controller.GetUserProfile)
	}

	required := r.Group("/").Use(a.AuthRequired)
	{
		required.GET("/user/:username/messages", a.WithUser(false), controller.GetUserMessages)
		required.POST("/user/:username/messages", a.WithUser(false), controller.PostMessage)

		required.GET("/", a.WithUser(true), controller.CheckCredentials)
		required.PATCH("/", a.WithUser(false), controller.PatchUser) // TODO: not working
		required.GET("/updates", a.WithUser(true), controller.GetUpdates)
	}
}
