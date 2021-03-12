package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/internal/app"
	"github.com/muerwre/vault-golang/internal/db"
	controller2 "github.com/muerwre/vault-golang/internal/feature/oauth/controller"
	"github.com/muerwre/vault-golang/pkg"
)

type OauthRouter struct {
	config     app.Config
	db         db.DB
	controller *controller2.OAuthController
	api        pkg.AppApi
}

func (or *OauthRouter) Init(api pkg.AppApi, db db.DB, config app.Config) *OauthRouter {
	or.db = db
	or.config = config
	or.controller = new(controller2.OAuthController).Init(db, config)
	or.api = api
	return or
}

func (or *OauthRouter) Handle(r *gin.RouterGroup) *OauthRouter {
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
		router.GET("/process/", or.controller.GetRedirectData(), or.controller.Process)

		router.DELETE("/:id", or.api.AuthRequired, or.api.WithUser(false), or.controller.Delete)
	}

	authenticated := r.Group("/", or.api.AuthRequired, or.api.WithUser(false))
	{
		authenticated.GET("/", or.controller.List)
	}

	return or
}
