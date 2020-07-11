package controllers

import (
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileController struct {
	db     db.DB
	config app.Config
}

func (fc *FileController) Init(db db.DB, config app.Config) {
	fc.db = db
	fc.config = config
}

func (fc *FileController) UploadFile(c *gin.Context) {
	user := c.MustGet("User").(*models.User)
	file, header, err := c.Request.FormFile("file")
	target := c.Param("target")
	fileType := c.Param("type")

	content := strings.Builder{}

	if _, err = io.Copy(&content, file); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.EmptyRequest})
		return
	}

	mime := mimetype.Detect([]byte(content.String()))
	inferredType := fc.GetFileType(mime.String())

	if inferredType == "" || inferredType != fileType {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.UnknownFileType})
		return
	}

	// TODO: check target

	path := fmt.Sprintf("%s/%d/%s", target, time.Now().Year(), time.Now().Month().String())
	cleanName := filepath.Base(filepath.Clean(header.Filename))
	fileExt := filepath.Ext(cleanName)
	fileName := cleanName[:len(cleanName)-len(fileExt)]
	name := fmt.Sprintf("%s-%d%s", fileName, time.Now().Unix(), fileExt)
	fullDir := fmt.Sprintf("%s/%s", filepath.Clean(fc.config.UploadPath), path)

	if err = os.MkdirAll(fullDir, os.ModePerm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData, "details": err.Error()})
		return
	}

	if out, err := os.Create(fmt.Sprintf("%s/%s", fullDir, name)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData, "details": err.Error()})
		return
	} else {
		defer out.Close()

		if _, err = out.WriteString(content.String()); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData, "details": err.Error()})
			return
		}
	}

	url := fmt.Sprintf("REMOTE_CURRENT://%s/%s", path, name)

	instance := &models.File{
		User:     user,
		Mime:     mime.String(),
		FullPath: fmt.Sprintf("%s/%s", path, name),
		Name:     name,
		Path:     path,
		OrigName: header.Filename,
		Url:      url,
		Size:     int(header.Size),
	}

	fmt.Printf("got file %+v", instance)

	c.JSON(http.StatusOK, gin.H{"file": instance})
	// TODO: check if it matches old api
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
