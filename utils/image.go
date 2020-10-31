package utils

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/gabriel-vasile/mimetype"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"github.com/lEx0/go-libjpeg-nrgba/jpeg"
	"github.com/muerwre/vault-golang/constants"
	"github.com/muerwre/vault-golang/utils/codes"
	"github.com/sirupsen/logrus"
	"image"
	"image/gif"
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
	constants.ImagePreset1600:      {Width: 1600},
	constants.ImagePreset600:       {Width: 600},
	constants.ImagePreset300:       {Width: 300},
	constants.ImagePresetAvatar:    {Width: 72, Height: 72, Crop: true},
	constants.ImagePresetCover:     {Width: 400, Height: 400, Crop: true},
	constants.ImagePresetSmallHero: {Width: 800, Height: 300, Crop: true},
}

func GetImagePresetByName(name string) *ImagePreset {
	for k, v := range ImagePresetList {
		if name == k {
			return v
		}
	}

	return nil
}

func WriteImageWebp(img image.Image, out io.Writer, mime string) (err error) {
	options, err := encoder.NewLossyEncoderOptions(encoder.PresetPhoto, 100)
	return webp.Encode(out, img, options)
}

func WriteImageInOriginalFormat(img image.Image, out io.Writer, mime string) (err error) {
	switch mime {
	case constants.FileMimeGif:
		err = gif.Encode(out, img, nil)
	case constants.FileMimeJpeg:
		err = jpeg.Encode(out, img, &jpeg.EncoderOptions{Quality: 100})
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
		*img, err = jpeg.Decode(file, &jpeg.DecoderOptions{})
	case constants.FileMimePng:
		*img, err = png.Decode(file)
	default:
		*img, err = nil, fmt.Errorf(codes.UnknownFileType)
	}

	return err
}

func CreateScaledImage(src string, dest string, presetName string, writeWebp bool) (*bytes.Buffer, error) {
	file, err := os.Open(src)

	if err != nil {
		logrus.Infof("Can't open file for cache transform: %s %s", src, err.Error())
		return nil, err
	}

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

	switch writeWebp {
	case true:
		// Write a file
		if err = WriteImageWebp(img, out, mime.String()); err != nil {
			return nil, err
		}

		// Write output
		if err = WriteImageWebp(img, content, mime.String()); err != nil {
			return nil, err
		}
	default:
		// Write a file
		if err = WriteImageInOriginalFormat(img, out, mime.String()); err != nil {
			return nil, err
		}

		// Write output
		if err = WriteImageInOriginalFormat(img, content, mime.String()); err != nil {
			return nil, err
		}
	}

	return content, err
}
