package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/db"
	controller2 "github.com/muerwre/vault-golang/feature/stats/controller"
	"github.com/muerwre/vault-golang/utils"
)

type StatsRouter struct {
	controller controller2.StatsController
	db         db.DB
}

func (sr *StatsRouter) Init(api utils.AppApi, db db.DB) {
	sr.controller = controller2.StatsController{DB: db}
	sr.db = db
}

// FlowRouter for /node/*
func (sr *StatsRouter) Handle(r *gin.RouterGroup) {
	r.GET("/", sr.controller.GetStats)
}
