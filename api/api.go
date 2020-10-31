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

	node   *routing.NodeRouter
	user   *routing.UserRouter
	stats  *routing.StatsRouter
	flow   *routing.FlowRouter
	upload *routing.UploadRouter
	static *routing.StaticRouter
	meta   *routing.MetaRouter
	oauth  *routing.OauthRouter
	search *routing.SearchRouter
	tag    *routing.TagRouter
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

	a.node = new(routing.NodeRouter).Init(a, a.db, a.Config).Handle(r.Group("/node"))
	a.user = new(routing.UserRouter).Init(a, a.db, a.mailer, a.Config).Handle(r.Group("/user"))
	a.search = new(routing.SearchRouter).Init(a, a.db).Handle(r.Group("/search"))
	a.oauth = new(routing.OauthRouter).Init(a, a.db, a.Config).Handle(r.Group("/oauth"))
	a.tag = new(routing.TagRouter).Init(a, a.db, a.Config).Handle(r.Group("/tag"))

	// TODO: do the same for:
	a.stats = &routing.StatsRouter{}
	a.stats.Init(a, a.db)

	a.flow = &routing.FlowRouter{}
	a.flow.Init(a, a.db, a.Config)

	a.upload = &routing.UploadRouter{}
	a.upload.Init(a, a.db, a.Config)

	a.static = &routing.StaticRouter{}
	a.static.Init(a, a.Config)

	a.meta = &routing.MetaRouter{}
	a.meta.Init(a.Config, a.db)

	a.Handle(r)

	return router
}

func (a *API) Handle(r *gin.RouterGroup) {
	a.stats.Handle(r.Group("/stats"))
	a.flow.Handle(r.Group("/flow"))
	a.upload.Handle(r.Group("/upload"))
	a.static.Handle(r.Group("/static"))
	a.meta.Handle(r.Group("/meta"))
}
