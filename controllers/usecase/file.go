package usecase

import (
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/constants"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils"
	"github.com/sirupsen/logrus"
	"image"
	"os"
	"path/filepath"
)

type FileUseCase struct {
	db     db.DB
	config app.Config
}

func (fu *FileUseCase) Init(db db.DB, config app.Config) *FileUseCase {
	fu.db = db
	fu.config = config
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
	path := filepath.Join(fu.config.UploadPath, f.Path, f.Name)

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
	case constants.FileTypeImage:
		if err := fu.FillMetadataImage(f); err != nil {
			logrus.Warnf("Can't get metadata for file %s: %s", f.Url, err.Error())
		}
		return
	case constants.FileTypeAudio:
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
			fu.db.File.UpdateMetadata(file, file.Metadata)
		}
	}

	return files
}
