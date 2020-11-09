package usecase

import (
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/db"
	fileConstants "github.com/muerwre/vault-golang/feature/file/constants"
	fileRepository "github.com/muerwre/vault-golang/feature/file/repository"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils"
	"github.com/muerwre/vault-golang/utils/codes"
	"github.com/sirupsen/logrus"
	"image"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type FileUseCase struct {
	config app.Config
	file   fileRepository.FileRepository
}

func (fu *FileUseCase) Init(db db.DB, config app.Config) *FileUseCase {
	fu.config = config
	fu.file = *new(fileRepository.FileRepository).Init(db.DB)
	return fu
}

// FillMetadataAudio fills Audio file metadata
func (fu FileUseCase) FillMetadataAudio(f *models.File) error {
	path := filepath.Join(fu.config.UploadPath, f.Path, f.Name)

	duration := utils.GetAudioDuration(path)
	artist, title := utils.GetAudioArtistTitle(path)

	if artist == "" && title == "" {
		title = f.OrigName
	}

	f.Metadata = models.FileMetadata{
		Id3title:  artist,
		Id3artist: title,
		Duration:  duration,
	}

	return nil
}

// FillMetadataImage fills Image file metadata
func (fu FileUseCase) FillMetadataImage(f *models.File) error {
	var path string

	file, err := os.Stat(filepath.Join(fu.config.UploadPath, f.Path))

	if err != nil {
		return err
	}

	switch mode := file.Mode(); {
	case mode.IsDir():
		path = filepath.Join(fu.config.UploadPath, f.Path, f.Name)
	case mode.IsRegular():
		path = filepath.Join(fu.config.UploadPath, f.Path)
	}

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

// FillMetadata fills file metadata
func (fu FileUseCase) FillMetadata(f *models.File) {
	switch f.Type {
	case fileConstants.FileTypeImage:
		if err := fu.FillMetadataImage(f); err != nil {
			logrus.Warnf("Can't get metadata for file %s: %s", f.Url, err.Error())
		}
		return
	case fileConstants.FileTypeAudio:
		if err := fu.FillMetadataAudio(f); err != nil {
			logrus.Warnf("Can't get metadata for file %s: %s", f.Url, err.Error())
		}
		return
	}
}

// UpdateFileMetadataIfNeeded fills node/comment files with proper metadata
func (fu FileUseCase) UpdateFileMetadataIfNeeded(files []*models.File) []*models.File {
	if len(files) == 0 {
		return files
	}

	for _, file := range files {
		if file.FileHasInvalidMetatada() {
			fu.FillMetadata(file)
			fu.file.UpdateMetadata(file, file.Metadata)
		}
	}

	return files
}

//
//func (fu FileUseCase) SaveFile(file *models.File) error {
//	return fu.file.Save(file)
//}

func (fu FileUseCase) CheckFileUploadSize(size int) error {
	if size > fu.config.UploadMaxSizeMb {
		return fmt.Errorf("file is too big for upload")
	}

	return nil
}

func (fu FileUseCase) CheckFileMimeAgainstUploadType(content []byte, fileType string) (mime string, err error) {
	mime = mimetype.Detect(content).String()
	inferredType := models.FileGetTypeByMime(mime)

	if inferredType == "" || inferredType != fileType {
		return "", fmt.Errorf(codes.UnknownFileType)
	}

	return mime, nil
}

func (fu FileUseCase) GenerateUploadFilename(name string, fileType string) (nameUnique string, fsFullDir string, pathCategorized string, err error) {
	year, month, _ := time.Now().Date()
	pathCategorized = filepath.Join("uploads", strconv.Itoa(year), strconv.Itoa(int(month)), fileType)
	cleanedSafeName := filepath.Base(filepath.Clean(name))
	fileExt := filepath.Ext(cleanedSafeName)
	fileName := cleanedSafeName[:len(cleanedSafeName)-len(fileExt)]

	nameUnique = fmt.Sprintf("%s-%d%s", fileName, time.Now().Unix(), fileExt)
	fsFullDir = filepath.Join(fu.config.UploadPath, pathCategorized)

	return nameUnique, fsFullDir, pathCategorized, nil
}

func (fu *FileUseCase) UploadRemotePic(url string, target string, fileType string, user *models.User) (*models.File, error) {
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	name := path.Base(url)

	result, err, _ := fu.SaveFile(
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

func (fu *FileUseCase) SaveFile(
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

	mime, err := fu.CheckFileMimeAgainstUploadType([]byte(content.String()), fileType)

	if err != nil {
		return nil, fmt.Errorf(codes.UnknownFileType), nil
	}

	if !models.FileValidateTarget(target) {
		return nil, fmt.Errorf(codes.IncorrectData), nil
	}

	nameUnique, fsFullDir, pathCategorized, err := fu.GenerateUploadFilename(name, fileType)
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

	fu.FillMetadata(&dbEntry)
	fu.file.Save(&dbEntry)

	return &dbEntry, nil, nil
}
