package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/internal/app"
	"github.com/muerwre/vault-golang/internal/db"
	controller2 "github.com/muerwre/vault-golang/internal/feature/tag/controller"
	"github.com/muerwre/vault-golang/pkg"
)

type TagRouter struct {
	controller controller2.TagController
	api        pkg.AppApi
}

func (r *TagRouter) Init(a pkg.AppApi, db db.DB, conf app.Config) *TagRouter {
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
