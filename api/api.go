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

	nodeRouter  *routing.NodeRouter
	userRouter  *routing.UserRouter
	statsRouter *routing.StatsRouter
	flowRouter  *routing.FlowRouter
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

func (a *API) Init(r *gin.RouterGroup) {
	r.Use(a.RecoverMiddleware, a.InjectContextMiddleware, a.OptionsRespondMiddleware)
	r.OPTIONS("/*path", a.CorsHandler)

	a.nodeRouter = &routing.NodeRouter{}
	a.nodeRouter.Init(a, a.db)

	a.userRouter = &routing.UserRouter{}
	a.userRouter.Init(a, a.db, a.mailer, a.Config)

	a.statsRouter = &routing.StatsRouter{}
	a.statsRouter.Init(a, a.db)

	a.flowRouter = &routing.FlowRouter{}
	a.flowRouter.Init(a, a.db)

	a.Handle(r)
}

func (a *API) Handle(r *gin.RouterGroup) {
	a.nodeRouter.Handle(r.Group("/node"))
	a.userRouter.Handle(r.Group("/user"))
	a.statsRouter.Handle(r.Group("/stats"))
	a.flowRouter.Handle(r.Group("/flow"))
}
