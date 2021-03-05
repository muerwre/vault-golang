package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/internal/db"
	"github.com/muerwre/vault-golang/internal/feature/notification/controller"
	"github.com/muerwre/vault-golang/pkg"
)

type NotificationRouter struct {
	api        pkg.AppApi
	controller controller.NotificationController
}

func (r *NotificationRouter) Init(api pkg.AppApi, db db.DB) *NotificationRouter {
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

	g.GET("/", r.api.AuthRequired, r.controller.GetNotifications)
	g.POST("/", r.api.AuthRequired, r.controller.PostSettings)

	return r
}
