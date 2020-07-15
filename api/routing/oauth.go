package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/controllers"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/utils"
)

type OauthRouter struct {
	config     app.Config
	db         db.DB
	controller *controllers.OAuthController
	api        utils.AppApi
}

func (or *OauthRouter) Init(api utils.AppApi, db db.DB, config app.Config) {
	or.db = db
	or.config = config
	or.controller = &controllers.OAuthController{DB: db, Config: config}
	or.api = api
	or.controller.Init()
}

func (or *OauthRouter) Handle(r *gin.RouterGroup) {
	router := r.Group("/:provider", or.controller.ProviderMiddleware)

	router.GET("/redirect", or.controller.Redirect)
	router.GET("/process", or.controller.Process)

	router.POST("/attach", or.api.AuthRequired, or.controller.Attach)
	router.POST("/login", or.api.AuthOptional, or.controller.Login)
}
