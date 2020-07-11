package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/utils"
)

type StatsRouter struct {
	api utils.AppApi
	db  db.DB
}

func (sr *StatsRouter) Init(api utils.AppApi, db db.DB) {
	sr.api = api
	sr.db = db
}

// FlowRouter for /node/*
func (sr *StatsRouter) Handle(r *gin.RouterGroup) {
	//r.GET("/", a.AuthOptional, controllers.Node.GetDiff)
}
