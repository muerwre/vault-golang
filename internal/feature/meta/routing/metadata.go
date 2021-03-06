package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/internal/app"
	"github.com/muerwre/vault-golang/internal/db"
	"github.com/muerwre/vault-golang/internal/feature/meta/controller"
)

type MetaRouter struct {
	controller controller.MetaController
}

func (mr *MetaRouter) Init(config app.Config, db db.DB) *MetaRouter {
	mr.controller = *new(controller.MetaController).Init(db, config)
	return mr
}

func (mr MetaRouter) Handle(r *gin.RouterGroup) {
	r.GET("/youtube", mr.controller.GetYoutubeTitles)
}
