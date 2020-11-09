package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/db"
	fileUsecase "github.com/muerwre/vault-golang/feature/file/usecase"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
)

type FileController struct {
	file fileUsecase.FileUseCase
}

func (fc *FileController) Init(db db.DB, config app.Config) *FileController {
	fc.file = *new(fileUsecase.FileUseCase).Init(db, config)
	return fc
}

func (fc *FileController) UploadFile(c *gin.Context) {
	user := c.MustGet("User").(*models.User)
	file, header, err := c.Request.FormFile("file")
	target := c.Param("target")
	fileType := c.Param("type")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	if err = fc.file.CheckFileUploadSize(int(header.Size)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.FilesIsTooBig})
		return
	}

	dbEntry, err, details := fc.file.SaveFile(file, target, fileType, header.Filename, user)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "details": details.Error()})
		return
	}

	c.JSON(http.StatusCreated, dbEntry)
}
