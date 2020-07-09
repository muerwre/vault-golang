package api

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/api/router"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/db"
)

type API struct {
	App        *app.App
	DB         *db.DB
	nodeRouter *router.NodeRouter
}

type ErrorCode struct {
	Code   string   `json:"code"`
	Stack  []string `json:"stack"`
	Reason string   `json:"reason"`
}

func New(a *app.App) (api *API, err error) {
	return &API{App: a, DB: a.DB}, nil
}

func (a *API) Init(r *gin.RouterGroup) {
	a.nodeRouter = &router.NodeRouter{}
	a.nodeRouter.Init(a, a.DB)

	r.Use(a.RecoverMiddleware, a.InjectContextMiddleware, a.OptionsRespondMiddleware)
	r.OPTIONS("/*path", a.CorsHandler)

	a.Handle(r)
}

func (a *API) Handle(r *gin.RouterGroup) {
	a.nodeRouter.Handle(r.Group("/node"))
	UserRouter(r.Group("/user"), a)
	FlowRouter(r.Group("/flow"), a)
	StatsRouter(r.Group("/stats"), a)
}
