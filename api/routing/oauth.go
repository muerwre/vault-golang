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
	r.POST(
		"/attach",
		or.api.AuthRequired,
		or.api.WithUser(false),
		or.controller.AttachConfirm,
	)

	r.POST("/login", or.controller.Login)

	router := r.Group("/:provider", or.controller.ProviderMiddleware)
	{
		router.GET("/redirect/", or.controller.Redirect)
		router.GET("/process/", or.controller.GetRedirectData(utils.ProcessTargetAttach), or.controller.Process)

		router.DELETE("/:id", or.api.AuthRequired, or.controller.Delete)
	}

	authenticated := r.Group("/", or.api.AuthRequired)
	{
		authenticated.GET("/", or.controller.List)
	}
}
