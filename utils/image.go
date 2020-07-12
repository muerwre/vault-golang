package utils

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/gabriel-vasile/mimetype"
	"github.com/muerwre/vault-golang/constants"
	"github.com/muerwre/vault-golang/utils/codes"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
)

type ImagePreset struct {
	Width  int
	Height int
	Crop   bool
}

var ImagePresetList = map[string]*ImagePreset{
	constants.ImagePreset1600:      &ImagePreset{Width: 1600},
	constants.ImagePreset600:       &ImagePreset{Width: 600},
	constants.ImagePreset300:       &ImagePreset{Width: 300},
	constants.ImagePresetAvatar:    &ImagePreset{Width: 72, Height: 72, Crop: true},
	constants.ImagePresetCover:     &ImagePreset{Width: 400, Height: 400, Crop: true},
	constants.ImagePresetSmallHero: &ImagePreset{Width: 800, Height: 300, Crop: true},
}

func GetImagePresetByName(name string) *ImagePreset {
	for k, v := range ImagePresetList {
		if name == k {
			return v
		}
	}

	return nil
}

func WriteImage(img image.Image, out io.Writer, mime string) (err error) {
	switch mime {
	case constants.FileMimeGif:
		err = gif.Encode(out, img, nil)
	case constants.FileMimeJpeg:
		err = jpeg.Encode(out, img, nil)
	case constants.FileMimePng:
		err = png.Encode(out, img)
	default:
		err = fmt.Errorf(codes.UnknownFileType)
	}

	return err
}

type AbstractOptions interface{}

func ReadImage(img *image.Image, file io.Reader, mime string) (err error) {
	switch mime {
	case constants.FileMimeGif:
		*img, err = gif.Decode(file)
	case constants.FileMimeJpeg:
		*img, err = jpeg.Decode(file)
	case constants.FileMimePng:
		*img, err = png.Decode(file)
	default:
		*img, err = nil, fmt.Errorf(codes.UnknownFileType)
	}

	return err
}

func CreateScaledImage(src string, dest string, presetName string) (*bytes.Buffer, error) {
	file, err := os.Open(src)

	mime, err := mimetype.DetectFile(src)

	if err != nil {
		return nil, err
	}

	var img image.Image = nil

	err = ReadImage(&img, file, mime.String())

	preset := GetImagePresetByName(presetName)

	if preset == nil {
		return nil, fmt.Errorf(codes.UnknownFileType)
	}

	switch preset.Crop {
	case true:
		img = imaging.Fill(img, preset.Width, preset.Height, imaging.Center, imaging.Lanczos)
	default:
		img = imaging.Resize(img, preset.Width, preset.Height, imaging.Lanczos)
	}

	if err = os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return nil, err
	}

	out, err := os.Create(dest)

	if err != nil {
		return nil, err
	}

	defer out.Close()

	content := bytes.NewBuffer([]byte{})

	if err = WriteImage(img, out, mime.String()); err != nil {
		return nil, err
	}

	if err = WriteImage(img, content, mime.String()); err != nil {
		return nil, err
	}

	return content, err
}
