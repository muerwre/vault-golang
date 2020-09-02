package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/controllers"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/utils"
)

type NotificationRouter struct {
	db         db.DB
	api        utils.AppApi
	controller controllers.NotificationController
}

func (r *NotificationRouter) Init(api utils.AppApi, db db.DB) *NotificationRouter {
	r.db = db
	r.api = api
	r.controller = *new(controllers.NotificationController).Init(db)

	return r
}

func (r *NotificationRouter) Handle(g *gin.RouterGroup) *NotificationRouter {
	node := g.Group("/node", r.api.AuthRequired)
	{
		node.GET("/:id", r.controller.NodeGet)
		node.POST("/:id/watch", r.controller.NodeWatch)
		node.POST("/:id/unwatch", r.controller.NodeUnwatch)
	}

	return r
}
