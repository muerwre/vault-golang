package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/internal/db"
	"github.com/muerwre/vault-golang/internal/feature/lab/controller"
	"github.com/muerwre/vault-golang/pkg"
)

type LabRouter struct {
	lab controller.LabController
	api pkg.AppApi
}

func (fr *LabRouter) Init(api pkg.AppApi, db db.DB) *LabRouter {
	fr.api = api
	fr.lab = *new(controller.LabController).Init(db)
	return fr
}

// LabRouter for /lab/*
func (fr *LabRouter) Handle(r *gin.RouterGroup) *LabRouter {
	r.GET("/", fr.lab.List)

	return fr
}
