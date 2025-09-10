package common

import (
	"fmt"
	"regexp"
)

func GetYouTubeID(url string) (string, error) {
	patterns := []string{
		`(?:v=|\/)([0-9A-Za-z_-]{11}).*`,
		`youtu\.be\/([0-9A-Za-z_-]{11}).*`,
	}
	for _, p := range patterns {
		re := regexp.MustCompile(p)
		match := re.FindStringSubmatch(url)
		if len(match) > 1 {
			return match[1], nil
		}
	}
	return "", fmt.Errorf("invalid YouTube URL")
}
