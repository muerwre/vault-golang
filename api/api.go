package api

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/api/routing"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/utils/mail"
)

type API struct {
	Config app.Config

	app    *app.App
	db     db.DB
	mailer mail.Mailer

	nodeRouter   *routing.NodeRouter
	userRouter   *routing.UserRouter
	statsRouter  *routing.StatsRouter
	flowRouter   *routing.FlowRouter
	uploadRouter *routing.UploadRouter
	staticRouter *routing.StaticRouter
	metaRouter   *routing.MetaRouter
	oauthRouter  *routing.OauthRouter
	searchRouter *routing.SearchRouter
}

// TODO: remove it? Or made it error response
type ErrorCode struct {
	Code   string   `json:"code"`
	Stack  []string `json:"stack"`
	Reason string   `json:"reason"`
}

func New(a *app.App) (api *API, err error) {
	return &API{app: a, db: *a.DB, Config: *a.Config, mailer: *a.Mailer}, nil
}

func (a *API) Init() *gin.Engine {
	var router *gin.Engine

	if !a.Config.ApiDebug {
		gin.SetMode(gin.ReleaseMode)
		router = gin.New()
	} else {
		router = gin.Default()
	}

	router.LoadHTMLGlob("templates/*")
	r := router.Group("/")
	r.Use(a.InjectContextMiddleware, a.OptionsRespondMiddleware)

	if !a.Config.Debug {
		r.Use(a.RecoverMiddleware)
	}

	r.OPTIONS("/*path", a.CorsHandler)

	a.nodeRouter = new(routing.NodeRouter).Init(a, a.db, a.Config).Handle(r.Group("/node"))
	a.userRouter = new(routing.UserRouter).Init(a, a.db, a.mailer, a.Config).Handle(r.Group("/user"))
	a.searchRouter = new(routing.SearchRouter).Init(a, a.db).Handle(r.Group("/search"))

	// TODO: do the same for:
	a.statsRouter = &routing.StatsRouter{}
	a.statsRouter.Init(a, a.db)

	a.flowRouter = &routing.FlowRouter{}
	a.flowRouter.Init(a, a.db, a.Config)

	a.uploadRouter = &routing.UploadRouter{}
	a.uploadRouter.Init(a, a.db, a.Config)

	a.staticRouter = &routing.StaticRouter{}
	a.staticRouter.Init(a, a.Config)

	a.metaRouter = &routing.MetaRouter{}
	a.metaRouter.Init(a.Config, a.db)

	a.oauthRouter = &routing.OauthRouter{}
	a.oauthRouter.Init(a, a.db, a.Config)

	a.Handle(r)

	return router
}

func (a *API) Handle(r *gin.RouterGroup) {
	a.statsRouter.Handle(r.Group("/stats"))
	a.flowRouter.Handle(r.Group("/flow"))
	a.uploadRouter.Handle(r.Group("/upload"))
	a.staticRouter.Handle(r.Group("/static"))
	a.metaRouter.Handle(r.Group("/meta"))
	a.oauthRouter.Handle(r.Group("/oauth"))
}
