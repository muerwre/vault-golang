package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/internal/app"
	"github.com/muerwre/vault-golang/internal/db"
	"github.com/muerwre/vault-golang/internal/feature/user/controller"
	"github.com/muerwre/vault-golang/internal/service/mail"
	"github.com/muerwre/vault-golang/pkg"
)

type UserRouter struct {
	controller controller.UserController
	api        pkg.AppApi
}

func (ur *UserRouter) Init(a pkg.AppApi, db db.DB, mailer mail.MailService, conf app.Config) *UserRouter {
	ur.controller = *new(controller.UserController).Init(db, mailer, conf)
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

		required.DELETE("/user/:username/messages/:id", ur.api.WithUser(false), ur.controller.DeleteMessage)

		required.GET("/", ur.api.WithUser(true), ur.controller.CheckCredentials)
		required.PATCH("/", ur.api.WithUser(false), ur.controller.PatchUser) // TODO: not working
		required.GET("/updates", ur.api.WithUser(true), ur.controller.GetUpdates)
	}

	return ur
}
