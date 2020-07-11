package controllers

import (
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
	"io"
	"net/http"
	"strings"
)

type FileController struct {
	db db.DB
}

func (fc *FileController) Init(db db.DB) {
	fc.db = db
}

func (fc *FileController) UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	target := c.Param("target")
	fileType := c.Param("type")

	content := strings.Builder{}
	_, err = io.Copy(&content, file)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.EmptyRequest})
		return
	}

	mime := mimetype.Detect([]byte(content.String()))
	inferredType := fc.GetFileType(mime.String())

	if inferredType == "" || inferredType != fileType {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.UnknownFileType})
		return
	}

	fmt.Printf("got file", file, header, err, target, fileType, mime, inferredType)

	c.JSON(http.StatusOK, gin.H{"file": file})
}

func (fc *FileController) GetFileType(fileMime string) string {
	for fileType, mimes := range models.FileTypeToMime {
		for _, mimeType := range mimes {
			if mimeType == fileMime {
				return fileType
			}
		}
	}

	return ""
}
