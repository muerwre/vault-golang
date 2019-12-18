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

	// r.POST("/login", controllers.User.LoginUser)
	// r.GET("/:username/profile", a.AuthOptional, controllers.User.GetUserProfile)

	// authorized := r.Group("/").Use(a.AuthRequired)
	// {
	// authorized.GET("/", a.WithUser(true), controllers.User.CheckCredentials)
	// authorized.PATCH("/", a.WithUser(false), controllers.User.PatchUser)
	// }
}
