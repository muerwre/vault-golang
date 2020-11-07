package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/db"
	controller2 "github.com/muerwre/vault-golang/feature/meta/controller"
)

type MetaRouter struct {
	config     app.Config
	controller controller2.MetaController
}

func (mr *MetaRouter) Init(config app.Config, db db.DB) {
	mr.config = config
	mr.controller = controller2.MetaController{Config: config, DB: db}
}

func (mr MetaRouter) Handle(r *gin.RouterGroup) {
	r.GET("/youtube", mr.controller.GetYoutubeTitles)
}
