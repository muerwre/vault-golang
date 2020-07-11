package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/controllers"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/utils"
)

type UploadRouter struct {
	controller controllers.FileController
	db         db.DB
	config     app.Config
}

func (ur *UploadRouter) Init(api utils.AppApi, db db.DB, config app.Config) {
	ur.controller = controllers.FileController{}
	ur.controller.Init(db)
	ur.db = db
	ur.config = config
}

// UploadRouter for /node/*
func (ur *UploadRouter) Handle(r *gin.RouterGroup) {
	r.POST("/:target/:type", ur.controller.UploadFile)
}
