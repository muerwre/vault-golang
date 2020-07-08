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

	r.GET("/user/:username/messages", a.AuthRequired, a.WithUser(false), controller.GetUserMessages)
	r.POST("/user/:username/messages", a.AuthRequired, a.WithUser(false), controller.PostMessage)
	r.GET("/user/:username/profile", a.AuthOptional, controller.GetUserProfile)

	required := r.Group("/").Use(a.AuthRequired)
	{
		required.GET("/", a.WithUser(true), controller.CheckCredentials)
		required.PATCH("/", a.WithUser(false), controller.PatchUser) // TODO: not working
		required.GET("/updates", a.WithUser(true), controller.GetUpdates)
	}
}
