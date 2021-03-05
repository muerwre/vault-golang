package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/internal/db"
	controller2 "github.com/muerwre/vault-golang/internal/feature/stats/controller"
	"github.com/muerwre/vault-golang/pkg"
)

type StatsRouter struct {
	controller controller2.StatsController
	db         db.DB
}

func (sr *StatsRouter) Init(api pkg.AppApi, db db.DB) {
	sr.controller = controller2.StatsController{DB: db}
	sr.db = db
}

// StatsRouter for /stats/*
func (sr *StatsRouter) Handle(r *gin.RouterGroup) {
	r.GET("/", sr.controller.GetStats)
}
