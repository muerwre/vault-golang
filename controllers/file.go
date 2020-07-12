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

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	if int(header.Size) > fc.config.UploadMaxSizeMb {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.FilesIsTooBig})
		return
	}

	content := strings.Builder{}

	if _, err = io.Copy(&content, file); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.EmptyRequest})
		return
	}

	mime := mimetype.Detect([]byte(content.String()))
	inferredType := models.FileGetTypeByMime(mime.String())

	// check type
	if inferredType == "" || inferredType != fileType {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.UnknownFileType})
		return
	}

	// TODO: check target

	pathCategorized := fmt.Sprintf("%s/%d/%s", target, time.Now().Year(), time.Now().Month().String())
	cleanedSafeName := filepath.Base(filepath.Clean(header.Filename))
	fileExt := filepath.Ext(cleanedSafeName)
	fileName := cleanedSafeName[:len(cleanedSafeName)-len(fileExt)]
	nameUnique := fmt.Sprintf("%s-%d%s", fileName, time.Now().Unix(), fileExt)
	fsFullDir := fmt.Sprintf("%s/%s", filepath.Clean(fc.config.UploadPath), pathCategorized)

	// recursively create destination folder
	if err = os.MkdirAll(fsFullDir, os.ModePerm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData, "details": err.Error()})
		return
	}

	// create dir and write file
	if out, err := os.Create(fmt.Sprintf("%s/%s", fsFullDir, nameUnique)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData, "details": err.Error()})
		return
	} else {
		defer out.Close()

		if _, err = out.WriteString(content.String()); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData, "details": err.Error()})
			return
		}
	}

	dbEntry := &models.File{
		User:     user,
		Mime:     mime.String(),
		FullPath: fmt.Sprintf("%s/%s", pathCategorized, nameUnique),
		Name:     nameUnique,
		Path:     pathCategorized,
		OrigName: header.Filename,
		Url:      fmt.Sprintf("REMOTE_CURRENT://%s/%s", pathCategorized, nameUnique),
		Size:     int(header.Size),
	}

	// TODO: save file at db

	c.JSON(http.StatusOK, gin.H{"file": dbEntry})
	// TODO: check if it matches old api
}
