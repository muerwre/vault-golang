package api

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/controllers"
)

// UserRouter for /node/*
func NodeRouter(r *gin.RouterGroup, a *API) {
	node := r.Group("/:id")
	{
		node.GET("/", a.AuthOptional, a.WithUser(false), controllers.Node.GetNode)
		node.POST("/tags", a.AuthRequired, a.WithUser(false), controllers.Node.PostTags)
		node.POST("/like", a.AuthRequired, a.WithUser(false), controllers.Node.PostLike)
		node.POST("/lock", a.AuthRequired, a.WithUser(false), controllers.Node.PostLock)
		node.POST("/heroic", a.AuthRequired, a.WithUser(false), controllers.Node.PostHeroic)
		node.POST("/cell-view", a.AuthRequired, a.WithUser(false), controllers.Node.PostCellView)
	}

	comment := r.Group("/:id/comment")
	{
		comment.GET("/", controllers.Node.GetNodeComments)
		comment.POST("/", a.AuthRequired, a.WithUser(false), controllers.Node.PostComment)
		comment.POST("/:cid/lock", a.AuthRequired, a.WithUser(false), controllers.Node.LockComment)
	}
}
