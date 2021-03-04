package utils

import (
	"fmt"
	"regexp"
)

var ThumbRegexps = [...]*regexp.Regexp{
	regexp.MustCompile(`(?:youtube\.com\/(?:[^\/]+\/.+\/|(?:v|e(?:mbed)?)\/|.*[?&]v=)|youtu\.be\/)([^"&?\/ ]{11})`),
}

var ThumbProviders = [...]string{
	"https://i.ytimg.com/vi/%s/hqdefault.jpg",
}

func GetThumbFromUrl(url string) string {
	for k, v := range ThumbRegexps {
		matches := v.FindSubmatch([]byte(url))

		if len(matches) > 0 && ThumbProviders[k] != "" {
			return fmt.Sprintf(ThumbProviders[k], string(matches[1]))
		}
	}

	return ""
}
