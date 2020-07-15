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

func (or *OauthRouter) Init(a utils.AppApi, db db.DB, config app.Config) {
	or.db = db
	or.config = config
	or.controller = &controllers.OAuthController{DB: db, Config: config}
}

func (or *OauthRouter) Handle(r *gin.RouterGroup) {
	router := r.Group("/:provider", or.controller.ProviderMiddleware)
	router.GET("/:provider/redirect", or.controller.Redirect)
	router.GET("/:provider/process", or.controller.Process)

	router.POST("/:provider/attach", or.api.AuthRequired, or.controller.Attach)
	router.POST("/:provider/login", or.api.AuthOptional, or.controller.Login)
}
