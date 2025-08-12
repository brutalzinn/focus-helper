package activity

import (
	"database/sql"
	"focus-helper/pkg/actions"
	"focus-helper/pkg/llm"
	"focus-helper/pkg/models"
	"focus-helper/pkg/notifications"
	"focus-helper/pkg/state"
	"focus-helper/pkg/variables"
	"strings"

	"github.com/go-vgo/robotgo"
)

type Monitor struct {
	lastMouseX int
	lastMouseY int
	deps       MonitorDependencies
}

type MonitorDependencies struct {
	AppConfig      *models.Config
	VarProcessor   *variables.Processor
	Notifier       notifications.Notifier
	LLMAdapter     llm.LLMAdapter
	AppState       *state.AppState
	DB             *sql.DB
	ActionExecutor *actions.Executor
}

func NewMonitor(deps MonitorDependencies) *Monitor {
	x, y := robotgo.Location()
	return &Monitor{lastMouseX: x, lastMouseY: y, deps: deps}
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

func DetectSubject(associations map[string]string) string {
	title := robotgo.GetTitle()
	if title == "" {
		return "Unknown"
	}
	for keyword, activity := range associations {
		if strings.Contains(strings.ToLower(title), strings.ToLower(keyword)) {
			return activity
		}
	}
	return "General Use"
}
