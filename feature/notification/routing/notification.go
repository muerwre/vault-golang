package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/feature/notification/controller"
	"github.com/muerwre/vault-golang/utils"
)

type NotificationRouter struct {
	api        utils.AppApi
	controller controller.NotificationController
}

func (r *NotificationRouter) Init(api utils.AppApi, db db.DB) *NotificationRouter {
	r.api = api
	r.controller = *new(controller.NotificationController).Init(db)
	return r
}

func (r *NotificationRouter) Handle(g *gin.RouterGroup) *NotificationRouter {
	node := g.Group("/node", r.api.AuthRequired)
	{
		node.GET("/:id", r.controller.NodeGet)
		node.POST("/:id", r.controller.NodePost)
		node.DELETE("/:id", r.controller.NodeDelete)
	}

	g.GET("/", r.api.AuthRequired, r.controller.GetSettings)
	g.POST("/", r.api.AuthRequired, r.controller.PostSettings)

	return r
}
