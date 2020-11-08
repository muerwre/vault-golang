package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/db"
	fileUsecase "github.com/muerwre/vault-golang/feature/file/usecase"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
	"github.com/sirupsen/logrus"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
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

	dbEntry, err, details := fc.SaveFile(file, target, fileType, header.Filename, user)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "details": details.Error()})
		return
	}

	c.JSON(http.StatusCreated, dbEntry)
}

func (fc *FileController) UploadRemotePic(url string, target string, fileType string, user *models.User) (*models.File, error) {
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	name := path.Base(url)

	result, err, _ := fc.SaveFile(
		resp.Body,
		target,
		fileType,
		name,
		user,
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (fc *FileController) SaveFile(
	reader io.Reader,
	target string,
	fileType string,
	name string,
	user *models.User,
) (result *models.File, error error, details error) {
	content := strings.Builder{}
	size, err := io.Copy(&content, reader)
	if err != nil {
		return nil, fmt.Errorf(codes.EmptyRequest), nil
	}

	mime, err := fc.file.CheckFileMimeAgainstUploadType([]byte(content.String()), fileType)

	if err != nil {
		return nil, fmt.Errorf(codes.UnknownFileType), nil
	}

	if !models.FileValidateTarget(target) {
		return nil, fmt.Errorf(codes.IncorrectData), nil
	}

	nameUnique, fsFullDir, pathCategorized, err := fc.file.GenerateUploadFilename(name, fileType)
	if err != nil {
		logrus.Infof("Error while uploding file %s: %s", name, err.Error())
		return nil, fmt.Errorf(codes.IncorrectData), nil
	}

	// recursively create destination folder
	if err := os.MkdirAll(fsFullDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf(codes.IncorrectData), err
	}

	// create dir and write file
	if out, err := os.Create(filepath.Join(fsFullDir, nameUnique)); err != nil {
		return nil, fmt.Errorf(codes.IncorrectData), err
	} else {
		defer out.Close()

		if _, err = out.WriteString(content.String()); err != nil {
			return nil, fmt.Errorf(codes.IncorrectData), err
		}
	}

	dbEntry := models.File{
		User:     user,
		Mime:     mime,
		FullPath: filepath.Join(pathCategorized, nameUnique),
		Name:     nameUnique,
		Path:     pathCategorized,
		OrigName: name,
		Url:      fmt.Sprintf("REMOTE_CURRENT://%s/%s", pathCategorized, nameUnique),
		Size:     int(size),
		Type:     fileType,
	}

	fc.file.FillMetadata(&dbEntry)
	fc.file.SaveFile(&dbEntry)

	return &dbEntry, nil, nil
}
