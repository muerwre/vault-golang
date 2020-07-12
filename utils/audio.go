package utils

import (
	"fmt"
	"github.com/goulash/audio"
	"github.com/tcolgate/mp3"
	"io"
	"os"
)

func GetAudioDuration(path string) int {
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
			fmt.Println(err)
			return 0
		}

		duration = duration + fr.Duration().Seconds()
	}

	return int(duration)
}

func GetAudioArtistTitle(path string) (artist string, title string) {
	metadata, err := audio.ReadMetadata(path)

	if err != nil {
		return "", ""
	}

	return metadata.Artist(), metadata.Title()
}