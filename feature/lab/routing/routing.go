package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/feature/lab/controller"
	"github.com/muerwre/vault-golang/utils"
)

type LabRouter struct {
	lab controller.LabController
	api utils.AppApi
}

func (fr *LabRouter) Init(api utils.AppApi, db db.DB) *LabRouter {
	fr.api = api
	fr.lab = *new(controller.LabController).Init(db)
	return fr
}

// LabRouter for /lab/*
func (fr *LabRouter) Handle(r *gin.RouterGroup) *LabRouter {
	r.GET("/", fr.lab.List)

	return fr
}
