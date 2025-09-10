package activity

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/go-vgo/robotgo"
)


type RobotInterface interface {
	Location() (int, int)
	GetTitle() string
}


type RobotWrapper struct{}

func (r RobotWrapper) Location() (int, int) {
	return robotgo.Location()
}

func (r RobotWrapper) GetTitle() string {
	return robotgo.GetTitle()
}


var robot RobotInterface = RobotWrapper{}


func safeGetTitle() string {
	title := ""
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from robotgo GetTitle panic: %v", r)
			title = "Unknown"
		}
	}()


	originalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w


	title = robot.GetTitle()


	w.Close()
	os.Stderr = originalStderr


	io.Copy(io.Discard, r)

	if title == "" {
		return "Unknown"
	}
	return title
}


func safeGetTitleWithRetry() string {
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		title := safeGetTitle()
		if title != "Unknown" {
			return title
		}

		time.Sleep(100 * time.Millisecond)
	}
	return "Unknown"
}
