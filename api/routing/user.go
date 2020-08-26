package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/controllers"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/utils"
	"github.com/muerwre/vault-golang/utils/mail"
)

type UserRouter struct {
	controller controllers.UserController
	api        utils.AppApi
}

func (ur *UserRouter) Init(a utils.AppApi, db db.DB, mailer mail.Mailer, conf app.Config) *UserRouter {
	ur.controller = *new(controllers.UserController).Init(db, mailer, conf)
	ur.api = a

	return ur
}

// UserRouter for /user/*
func (ur *UserRouter) Handle(r *gin.RouterGroup) *UserRouter {
	r.POST("/login", ur.controller.LoginUser)

	r.GET("/restore/:id", ur.controller.GetRestoreCode)
	r.POST("/restore/:id", ur.controller.PostRestoreCode)
	r.POST("/restore/", ur.controller.CreateRestoreCode)

	optional := r.Group("/").Use(ur.api.AuthOptional)
	{
		optional.GET("/user/:username/profile", ur.controller.GetUserProfile)
	}

	required := r.Group("/").Use(ur.api.AuthRequired)
	{
		required.GET("/user/:username/messages", ur.api.WithUser(false), ur.controller.GetUserMessages)
		required.POST("/user/:username/messages", ur.api.WithUser(false), ur.controller.PostMessage)

		required.GET("/", ur.api.WithUser(true), ur.controller.CheckCredentials)
		required.PATCH("/", ur.api.WithUser(false), ur.controller.PatchUser) // TODO: not working
		required.GET("/updates", ur.api.WithUser(true), ur.controller.GetUpdates)
	}

	return ur
}
