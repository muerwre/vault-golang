package router

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/controllers"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/utils"
)

type NodeRouter struct {
	controller *controllers.NodeController
	api        utils.AppApi
}

func (nr *NodeRouter) Init(a utils.AppApi, db *db.DB) {
	nr.controller = &controllers.NodeController{db}
	nr.api = a
}

func (nr *NodeRouter) Handle(r *gin.RouterGroup) {
	a := nr.api
	controller := nr.controller

	r.POST("/", a.AuthRequired, a.WithUser(false), controller.PostNode)

	node := r.Group("/:id")
	{
		node.GET("", a.AuthOptional, a.WithUser(false), controller.GetNode) // TODO: problem with files
		node.GET("/related", controller.GetRelated)

		node.POST("/tags", a.AuthRequired, a.WithUser(false), controller.PostTags)
		node.POST("/like", a.AuthRequired, a.WithUser(false), controller.PostLike)
		node.POST("/lock", a.AuthRequired, a.WithUser(false), controller.PostLock)
		node.POST("/heroic", a.AuthRequired, a.WithUser(false), controller.PostHeroic)
		node.POST("/cell-view", a.AuthRequired, a.WithUser(false), controller.PostCellView)
	}

	comment := r.Group("/:id/comment")
	{
		comment.GET("", controller.GetNodeComments)
		comment.POST("", a.AuthRequired, a.WithUser(false), controller.PostComment)
		comment.POST("/:cid/lock", a.AuthRequired, a.WithUser(false), controller.LockComment)
	}
}
