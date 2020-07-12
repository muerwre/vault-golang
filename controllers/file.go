package controllers

import (
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/goulash/audio"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

func (fc *FileController) FillMetadataAudio(f *models.File) error {
	path := filepath.Join(fc.config.UploadPath, f.FullPath)

	metadata, err := audio.ReadMetadata(path)

	if err != nil {
		return err
	}

	f.Metadata = models.FileMetadata{
		Id3title:  metadata.Title(),
		Id3artist: metadata.Artist(),
		Duration:  int(metadata.Length().Seconds()),
	}

	return nil
}

func (fc *FileController) FillMetadataImage(f *models.File) error {
	path := filepath.Join(fc.config.UploadPath, f.FullPath)

	if reader, err := os.Open(path); err == nil {
		defer reader.Close()
		im, _, err := image.DecodeConfig(reader)

		if err != nil {
			return err
		}

		f.Metadata = models.FileMetadata{
			Width:  im.Width,
			Height: im.Height,
		}

		return nil
	} else {
		return err
	}
}

func (fc *FileController) FillMetadata(f *models.File) {
	switch f.Type {
	case models.FileTypeImage:
		err := fc.FillMetadataImage(f)
		println("%+v", err)
		return
	case models.FileTypeAudio:
		_ = fc.FillMetadataAudio(f)
		return
	}
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

	if !models.FileValidateTarget(target) {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	pathCategorized := filepath.Join(target, strconv.Itoa(time.Now().Year()), time.Now().Month().String())
	cleanedSafeName := filepath.Base(filepath.Clean(header.Filename))
	fileExt := filepath.Ext(cleanedSafeName)
	fileName := cleanedSafeName[:len(cleanedSafeName)-len(fileExt)]
	nameUnique := fmt.Sprintf("%s-%d%s", fileName, time.Now().Unix(), fileExt)
	fsFullDir := filepath.Join(fc.config.UploadPath, pathCategorized)

	// recursively create destination folder
	if err = os.MkdirAll(fsFullDir, os.ModePerm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData, "details": err.Error()})
		return
	}

	// create dir and write file
	if out, err := os.Create(filepath.Join(fsFullDir, nameUnique)); err != nil {
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
		FullPath: filepath.Join(pathCategorized, nameUnique),
		Name:     nameUnique,
		Path:     pathCategorized,
		OrigName: header.Filename,
		Url:      fmt.Sprintf("REMOTE_CURRENT://%s/%s", pathCategorized, nameUnique),
		Size:     int(header.Size),
		Type:     fileType,
	}

	fc.FillMetadata(dbEntry)

	fc.db.FileRepository.Save(dbEntry)

	c.JSON(http.StatusOK, gin.H{"file": dbEntry})
	// TODO: check if it matches old api
}
