package activity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type FakeRobot struct {
	X, Y  int
	Title string
}

func (f *FakeRobot) Location() (int, int) { return f.X, f.Y }
func (f *FakeRobot) GetTitle() string     { return f.Title }

func TestHasActivity(t *testing.T) {

	originalRobot := robot
	defer func() { robot = originalRobot }()

	fake := &FakeRobot{X: 100, Y: 200}
	robot = fake
	act := &Activity{lastMouseX: 100, lastMouseY: 200}
	assert.False(t, act.HasActivity())
	fake.X = 150
	assert.True(t, act.HasActivity())
}

func TestDetectSubject(t *testing.T) {

	originalRobot := robot
	defer func() { robot = originalRobot }()

	associations := map[string]string{
		"chrome": "Browsing",
		"word":   "Writing",
	}
	robot = &FakeRobot{Title: "Google Chrome - FocusHelper"}
	subject := DetectSubject(associations)
	assert.Equal(t, "Browsing", subject)

	robot = &FakeRobot{Title: "Some Unknown App"}
	subject = DetectSubject(associations)
	assert.Equal(t, "General Use", subject)
}

func TestGetMainSubject(t *testing.T) {
	associations := map[string]string{
		"chrome": "Browsing",
		"word":   "Writing",
	}


	subject := GetMainSubject(map[string]int{}, associations)
	assert.Equal(t, "General Use", subject)


	frequency := map[string]int{
		"Browsing": 5,
		"Writing":  3,
		"Gaming":   1,
	}
	subject = GetMainSubject(frequency, associations)
	assert.Equal(t, "Browsing", subject)
}
