package utils

import (
	"github.com/goulash/audio"
	"github.com/sirupsen/logrus"
	"github.com/tcolgate/mp3"
	"io"
	"os"
)

func GetAudioDurationFromPath(path string) int {
	r, err := os.Open(path)
	duration := 0.0

	if err != nil {
		return 0
	}

	d := mp3.NewDecoder(r)
	var fr mp3.Frame
	skipped := 0

	for {
		if err := d.Decode(&fr, &skipped); err != nil {
			if err == io.EOF {
				break
			}

			logrus.Warnf(err.Error())
			return 0
		}

		duration = duration + fr.Duration().Seconds()
	}

	return int(duration)
}

func GetAudioArtistTitleFromPath(path string) (artist string, title string) {
	metadata, err := audio.ReadMetadata(path)

	if err != nil {
		return "", ""
	}

	return metadata.Artist(), metadata.Title()
}
