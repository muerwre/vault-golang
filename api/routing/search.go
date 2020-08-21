package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/controllers"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/utils"
)

type SearchRouter struct {
	api utils.AppApi
	db  db.DB

	controller *controllers.SearchController
}

func (sr *SearchRouter) Init(a utils.AppApi, db db.DB) *SearchRouter {
	sr.api = a
	sr.db = db
	sr.controller = new(controllers.SearchController).Init(db)

	return sr
}

func (sr *SearchRouter) Handle(r *gin.RouterGroup) *SearchRouter {
	controller := sr.controller

	r.GET("/nodes", controller.SearchNodes)

	return sr
}
