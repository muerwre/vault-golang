package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/controllers"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/utils"
)

type StatsRouter struct {
	controller controllers.StatsController
	db         db.DB
}

func (sr *StatsRouter) Init(api utils.AppApi, db db.DB) {
	sr.controller = controllers.StatsController{DB: db}
	sr.db = db
}

// FlowRouter for /node/*
func (sr *StatsRouter) Handle(r *gin.RouterGroup) {
	r.GET("/", sr.controller.GetStats)
}
