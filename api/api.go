package api

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/db"
	flowRouting "github.com/muerwre/vault-golang/feature/flow/routing"
	labRouting "github.com/muerwre/vault-golang/feature/lab/routing"
	metaRouting "github.com/muerwre/vault-golang/feature/meta/routing"
	nodeRouting "github.com/muerwre/vault-golang/feature/node/routing"
	notificationRouting "github.com/muerwre/vault-golang/feature/notification/routing"
	oauthRouting "github.com/muerwre/vault-golang/feature/oauth/routing"
	searchRouting "github.com/muerwre/vault-golang/feature/search/routing"
	staticRouting "github.com/muerwre/vault-golang/feature/static/routing"
	statsRouting "github.com/muerwre/vault-golang/feature/stats/routing"
	tagRouting "github.com/muerwre/vault-golang/feature/tag/routing"
	uploadRouting "github.com/muerwre/vault-golang/feature/upload/routing"
	userRouting "github.com/muerwre/vault-golang/feature/user/routing"
	"github.com/muerwre/vault-golang/service/mail"
	"github.com/muerwre/vault-golang/service/notification/controller"
)

type API struct {
	Config app.Config

	app      *app.App
	db       db.DB
	mailer   mail.MailService
	notifier controller.NotificationService

	node               *nodeRouting.NodeRouter
	user               *userRouting.UserRouter
	stats              *statsRouting.StatsRouter
	flow               *flowRouting.FlowRouter
	upload             *uploadRouting.UploadRouter
	static             *staticRouting.StaticRouter
	meta               *metaRouting.MetaRouter
	oauth              *oauthRouting.OauthRouter
	search             *searchRouting.SearchRouter
	tag                *tagRouting.TagRouter
	notificationRouter *notificationRouting.NotificationRouter
	lab                *labRouting.LabRouter
}

// TODO: remove it? Or made it error response
type ErrorCode struct {
	Code   string   `json:"code"`
	Stack  []string `json:"stack"`
	Reason string   `json:"reason"`
}

func New(a *app.App) (api *API, err error) {
	return &API{
		app:      a,
		db:       *a.DB,
		Config:   *a.Config,
		mailer:   *a.Mailer,
		notifier: *a.Notifier,
	}, nil
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

	a.node = new(nodeRouting.NodeRouter).Init(a, a.db, a.Config, a.notifier).Handle(r.Group("/node"))
	a.user = new(userRouting.UserRouter).Init(a, a.db, a.mailer, a.Config).Handle(r.Group("/user"))
	a.search = new(searchRouting.SearchRouter).Init(a, a.db).Handle(r.Group("/search"))
	a.oauth = new(oauthRouting.OauthRouter).Init(a, a.db, a.Config).Handle(r.Group("/oauth"))
	a.tag = new(tagRouting.TagRouter).Init(a, a.db, a.Config).Handle(r.Group("/tag"))
	a.notificationRouter = new(notificationRouting.NotificationRouter).Init(a, a.db).Handle(r.Group("/notifications"))
	a.lab = new(labRouting.LabRouter).Init(a, a.db).Handle(r.Group("/lab"))

	// TODO: do the same for:
	a.stats = &statsRouting.StatsRouter{}
	a.stats.Init(a, a.db)

	a.flow = &flowRouting.FlowRouter{}
	a.flow.Init(a, a.db, a.Config, a.notifier)

	a.upload = &uploadRouting.UploadRouter{}
	a.upload.Init(a, a.db, a.Config)

	a.static = &staticRouting.StaticRouter{}
	a.static.Init(a, a.Config)

	a.meta = &metaRouting.MetaRouter{}
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
