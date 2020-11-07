package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/db"
	controller2 "github.com/muerwre/vault-golang/feature/search/controller"
	"github.com/muerwre/vault-golang/utils"
)

type SearchRouter struct {
	api utils.AppApi
	db  db.DB

	controller *controller2.SearchController
}

func (sr *SearchRouter) Init(a utils.AppApi, db db.DB) *SearchRouter {
	sr.api = a
	sr.db = db
	sr.controller = new(controller2.SearchController).Init(db)

	return sr
}

func (sr *SearchRouter) Handle(r *gin.RouterGroup) *SearchRouter {
	controller := sr.controller

	r.GET("/nodes", controller.SearchNodes)

	return sr
}
