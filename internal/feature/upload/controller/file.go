package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/internal/app"
	"github.com/muerwre/vault-golang/internal/db"
	"github.com/muerwre/vault-golang/internal/db/models"
	fileUsecase "github.com/muerwre/vault-golang/internal/feature/upload/usecase"
	"github.com/muerwre/vault-golang/pkg/codes"
	"github.com/sirupsen/logrus"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
)

type UploadController struct {
	file fileUsecase.FileUseCase
}

func (uc *UploadController) Init(db db.DB, config app.Config) *UploadController {
	uc.file = *new(fileUsecase.FileUseCase).Init(db, config)
	return uc
}

func (uc *UploadController) UploadFile(c *gin.Context) {
	user := c.MustGet("User").(*models.User)
	file, header, err := c.Request.FormFile("file")
	target := c.Param("target")
	fileType := c.Param("type")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	if err = uc.file.CheckFileUploadSize(int(header.Size)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.FilesIsTooBig})
		return
	}

	dbEntry, err := uc.file.SaveFile(file, target, fileType, header.Filename, user)
	if err != nil {
		logrus.Warnf("Can't upload file: %s\nFile: %+v", err.Error(), file)
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.UnknownFileType})
		return
	}

	c.JSON(http.StatusCreated, dbEntry)
}
