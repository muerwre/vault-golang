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
	}

	comment := r.Group("/:id/comment")
	{
		comment.GET("/", controllers.Node.GetNodeComments)
	}
}
