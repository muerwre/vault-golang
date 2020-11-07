package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/db"
	controller2 "github.com/muerwre/vault-golang/feature/tag/controller"
	"github.com/muerwre/vault-golang/utils"
)

type TagRouter struct {
	controller controller2.TagController
	api        utils.AppApi
}

func (r *TagRouter) Init(a utils.AppApi, db db.DB, conf app.Config) *TagRouter {
	r.controller = *new(controller2.TagController).Init(db, conf)
	r.api = a

	return r
}

// TagRouter for /tag/*
func (r *TagRouter) Handle(rg *gin.RouterGroup) *TagRouter {
	rg.GET("/nodes", r.controller.GetNodesOfTag)
	rg.GET("/autocomplete", r.controller.GetAutocomplete)
	return r
}
