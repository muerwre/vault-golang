package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/controllers"
	"github.com/muerwre/vault-golang/db"
)

type MetaRouter struct {
	config     app.Config
	controller controllers.MetaController
}

func (mr *MetaRouter) Init(config app.Config, db db.DB) {
	mr.config = config
	mr.controller = controllers.MetaController{Config: config, DB: db}
}

func (mr MetaRouter) Handle(r *gin.RouterGroup) {
	r.GET("/youtube", mr.controller.GetYoutubeTitles)
}
