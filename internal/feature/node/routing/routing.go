package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/internal/app"
	"github.com/muerwre/vault-golang/internal/db"
	"github.com/muerwre/vault-golang/internal/feature/node/controller"
	controller2 "github.com/muerwre/vault-golang/internal/service/notification/controller"
	"github.com/muerwre/vault-golang/pkg"
)

type NodeRouter struct {
	controller *controller.NodeController
	api        pkg.AppApi
}

func (nr *NodeRouter) Init(a pkg.AppApi, db db.DB, config app.Config, notifier controller2.NotificationService) *NodeRouter {
	nr.controller = new(controller.NodeController).Init(db, config, notifier)
	nr.api = a
	return nr
}

func (nr *NodeRouter) Handle(r *gin.RouterGroup) *NodeRouter {
	a := nr.api
	controller := nr.controller

	r.POST("/", a.AuthRequired, a.WithUser(true), controller.PostNode)

	node := r.Group("/:id")
	{
		node.GET("", a.AuthOptional, a.WithUser(false), controller.GetNode)
		node.GET("/related", controller.GetRelated)

		node.POST("/tags", a.AuthRequired, a.WithUser(false), controller.PostTags)
		node.POST("/like", a.AuthRequired, a.WithUser(false), controller.PostLike)
		node.POST("/lock", a.AuthRequired, a.WithUser(false), controller.LockNode)
		node.POST("/heroic", a.AuthRequired, a.WithUser(false), controller.PostHeroic)
		node.POST("/cell-view", a.AuthRequired, a.WithUser(false), controller.PostCellView)
	}

	comment := r.Group("/:id/comment")
	{
		comment.GET("", controller.GetNodeComments)
		comment.POST("", a.AuthRequired, a.WithUser(true), controller.PostComment)
		comment.POST("/:cid/lock", a.AuthRequired, a.WithUser(false), controller.LockComment)
	}

	return nr
}
