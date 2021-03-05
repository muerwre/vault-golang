package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/internal/app"
	"github.com/muerwre/vault-golang/internal/db"
	controller2 "github.com/muerwre/vault-golang/internal/feature/upload/controller"
	"github.com/muerwre/vault-golang/pkg"
)

type UploadRouter struct {
	controller *controller2.UploadController
	db         db.DB
	config     app.Config
	api        pkg.AppApi
}

func (ur *UploadRouter) Init(api pkg.AppApi, db db.DB, config app.Config) {
	ur.controller = new(controller2.UploadController).Init(db, config)
	ur.api = api
	ur.db = db
	ur.config = config
}

// UploadRouter for /node/*
func (ur *UploadRouter) Handle(r *gin.RouterGroup) {
	r.POST("/:target/:type", ur.api.AuthRequired, ur.api.WithUser(false), ur.controller.UploadFile)
}
