package activity

import "github.com/go-vgo/robotgo"

type Monitor struct {
	lastMouseX int
	lastMouseY int
}

func NewMonitor() *Monitor {
	x, y := robotgo.Location()
	return &Monitor{lastMouseX: x, lastMouseY: y}
}

func (m *Monitor) HasActivity() bool {
	currentX, currentY := robotgo.Location()
	if currentX != m.lastMouseX || currentY != m.lastMouseY {
		m.lastMouseX = currentX
		m.lastMouseY = currentY
		return true
	}
	return false
}
