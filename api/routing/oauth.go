package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/controllers"
	"github.com/muerwre/vault-golang/db"
)

type OauthRouter struct {
	config     app.Config
	db         db.DB
	controller *controllers.OauthController
}

func (or *OauthRouter) Init(db db.DB, config app.Config) {
	or.db = db
	or.config = config
	or.controller = &controllers.OauthController{DB: db, Config: config}
}

func (or *OauthRouter) Handle(r *gin.RouterGroup) {
	r.GET("/vk/login", or.controller.RedirectVK)
	r.GET("/vk/process", or.controller.ProcessVkLogin)
	r.GET("/google/login", or.controller.RedirectGoogle)
	r.GET("/google/process", or.controller.ProcessGoogleLogin)
}
