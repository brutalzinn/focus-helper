package activity

import (
	"strings"
)


func DetectSubject(associations map[string]string) string {
	title := safeGetTitleWithRetry()
	if title == "" || title == "Unknown" {
		return "Unknown"
	}
	for keyword, activity := range associations {
		if strings.Contains(strings.ToLower(title), strings.ToLower(keyword)) {
			return activity
		}
	}
	return "General Use"
}


func GetMainSubject(subjectFrequency map[string]int, associations map[string]string) string {
	if len(subjectFrequency) == 0 {
		return DetectSubject(associations)
	}
	mainSubject := "Unknown"
	maxCount := 0
	for subject, count := range subjectFrequency {
		if count > maxCount {
			maxCount = count
			mainSubject = subject
		}
	}
	return mainSubject
}
