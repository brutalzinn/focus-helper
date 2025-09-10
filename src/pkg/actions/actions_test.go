package actions

import (
	"context"
	"focus-helper/src/pkg/models"
	"focus-helper/src/pkg/state"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	appState := &state.AppState{}
	Init(appState)

	if currentAppState != appState {
		t.Error("Init() did not set currentAppState correctly")
	}
}

func TestExecute_InvalidAction(t *testing.T) {
	appState := &state.AppState{
		EventChannel: make(chan state.AppEvent, 10),
	}
	Init(appState)

	action := models.ActionConfig{
		Type: "invalid_action",
	}

	err := Execute(action)
	if err == nil {
		t.Error("Expected error for invalid action type")
	}

	expectedError := "wrong action: invalid_action"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestExecute_StopAction(t *testing.T) {
	appState := &state.AppState{
		EventChannel: make(chan state.AppEvent, 10),
	}
	Init(appState)

	action := models.ActionConfig{
		Type: models.ActionStop,
	}

	err := Execute(action)
	if err != nil {
		t.Errorf("Stop action should not return error, got: %v", err)
	}
}

func TestExecute_PopupAction(t *testing.T) {
	appState := &state.AppState{
		EventChannel: make(chan state.AppEvent, 10),
		Notifier:     &mockNotifier{},
	}
	Init(appState)

	action := models.ActionConfig{
		Type:         models.ActionPopup,
		PopupTitle:   "Test Title",
		PopupMessage: "Test Message",
	}

	err := Execute(action)
	if err != nil {
		t.Errorf("Popup action should not return error, got: %v", err)
	}
}

func TestExecuteSequence_EmptySequence(t *testing.T) {
	appState := &state.AppState{
		EventChannel: make(chan state.AppEvent, 10),
	}
	Init(appState)

	actions := []models.ActionConfig{}

	ExecuteSequence(actions)
}

func TestExecuteSequence_SingleAction(t *testing.T) {
	appState := &state.AppState{
		EventChannel: make(chan state.AppEvent, 10),
		Notifier:     &mockNotifier{},
	}
	Init(appState)

	actions := []models.ActionConfig{
		{
			Type:         models.ActionPopup,
			PopupTitle:   "Test",
			PopupMessage: "Test Message",
		},
	}

	ExecuteSequence(actions)
}

func TestExecuteSequence_MultipleActions(t *testing.T) {
	appState := &state.AppState{
		EventChannel: make(chan state.AppEvent, 10),
		Notifier:     &mockNotifier{},
	}
	Init(appState)

	actions := []models.ActionConfig{
		{
			Type:         models.ActionPopup,
			PopupTitle:   "Test 1",
			PopupMessage: "Test Message 1",
		},
		{
			Type:         models.ActionPopup,
			PopupTitle:   "Test 2",
			PopupMessage: "Test Message 2",
		},
	}

	ExecuteSequence(actions)
}

func TestExecuteSequence_Cancellation(t *testing.T) {
	appState := &state.AppState{
		EventChannel: make(chan state.AppEvent, 10),
		Notifier:     &mockNotifier{},
	}
	Init(appState)

	actions := []models.ActionConfig{
		{
			Type:         models.ActionPopup,
			PopupTitle:   "Test 1",
			PopupMessage: "Test Message 1",
		},
		{
			Type:         models.ActionPopup,
			PopupTitle:   "Test 2",
			PopupMessage: "Test Message 2",
		},
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		StopCurrentActions()
	}()

	ExecuteSequence(actions)
}

func TestStopCurrentActions(t *testing.T) {
	appState := &state.AppState{
		EventChannel: make(chan state.AppEvent, 10),
	}
	Init(appState)

	err := StopCurrentActions()
	if err != nil {
		t.Errorf("StopCurrentActions should not return error, got: %v", err)
	}
}

func TestExecute_ConcurrentAccess(t *testing.T) {
	appState := &state.AppState{
		EventChannel: make(chan state.AppEvent, 100),
		Notifier:     &mockNotifier{},
	}
	Init(appState)

	action := models.ActionConfig{
		Type:         models.ActionPopup,
		PopupTitle:   "Test",
		PopupMessage: "Test Message",
	}

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				err := Execute(action)
				if err != nil {
					t.Errorf("Concurrent execution failed: %v", err)
				}
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestExecute_EventChannel(t *testing.T) {
	appState := &state.AppState{
		EventChannel: make(chan state.AppEvent, 10),
		Notifier:     &mockNotifier{},
	}
	Init(appState)

	action := models.ActionConfig{
		Type:         models.ActionPopup,
		PopupTitle:   "Test",
		PopupMessage: "Test Message",
	}

	events := make([]state.AppEvent, 0)
	go func() {
		for event := range appState.EventChannel {
			events = append(events, event)
		}
	}()

	err := Execute(action)
	if err != nil {
		t.Errorf("Execute should not return error, got: %v", err)
	}

	time.Sleep(10 * time.Millisecond)
	close(appState.EventChannel)

	if len(events) < 2 {
		t.Errorf("Expected at least 2 events (STOP_LISTENING and START_LISTENING), got %d", len(events))
	}

	if events[0].Type != "STOP_LISTENING" {
		t.Errorf("Expected first event to be STOP_LISTENING, got %s", events[0].Type)
	}

	if events[1].Type != "START_LISTENING" {
		t.Errorf("Expected second event to be START_LISTENING, got %s", events[1].Type)
	}
}

func TestExecute_ContextCancellation(t *testing.T) {
	appState := &state.AppState{
		EventChannel: make(chan state.AppEvent, 10),
		Notifier:     &mockNotifier{},
	}
	Init(appState)

	action := models.ActionConfig{
		Type:         models.ActionPopup,
		PopupTitle:   "Test",
		PopupMessage: "Test Message",
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	actionCancelFn = cancel

	err := Execute(action)
	if err != nil {
		t.Errorf("Execute should not return error for popup action, got: %v", err)
	}
}

func TestExecuteSequence_ContextCancellation(t *testing.T) {
	appState := &state.AppState{
		EventChannel: make(chan state.AppEvent, 10),
		Notifier:     &mockNotifier{},
	}
	Init(appState)

	actions := []models.ActionConfig{
		{
			Type:         models.ActionPopup,
			PopupTitle:   "Test 1",
			PopupMessage: "Test Message 1",
		},
		{
			Type:         models.ActionPopup,
			PopupTitle:   "Test 2",
			PopupMessage: "Test Message 2",
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	sequenceCancelFn = cancel

	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	ExecuteSequence(actions)
}

func TestExecute_ActionTypes(t *testing.T) {
	appState := &state.AppState{
		EventChannel: make(chan state.AppEvent, 10),
		Notifier:     &mockNotifier{},
	}
	Init(appState)

	testCases := []struct {
		name   string
		action models.ActionConfig
	}{
		{
			name: "Sound Action",
			action: models.ActionConfig{
				Type:      models.ActionSound,
				SoundFile: "test.wav",
				Volume:    0.5,
			},
		},
		{
			name: "Speak Action",
			action: models.ActionConfig{
				Type: models.ActionSpeak,
				Text: "Test message",
			},
		},
		{
			name: "SpeakIA Action",
			action: models.ActionConfig{
				Type:   models.ActionSpeakIA,
				Prompt: "Test prompt",
			},
		},
		{
			name: "YouTube Audio Action",
			action: models.ActionConfig{
				Type:    models.ActionYoutubeAudio,
				URL:     "https://www.youtube.com/watch?v=test",
				StartAt: "0:10",
				EndAt:   "0:20",
			},
		},
		{
			name: "Popup Action",
			action: models.ActionConfig{
				Type:         models.ActionPopup,
				PopupTitle:   "Test Title",
				PopupMessage: "Test Message",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Execute(tc.action)
			if err != nil {
				t.Errorf("Action %s failed: %v", tc.name, err)
			}
		})
	}
}

func TestExecute_VolumeDefault(t *testing.T) {
	appState := &state.AppState{
		EventChannel: make(chan state.AppEvent, 10),
	}
	Init(appState)

	action := models.ActionConfig{
		Type:      models.ActionSound,
		SoundFile: "test.wav",
		Volume:    0,
	}

	err := Execute(action)
	if err != nil {
		t.Errorf("Sound action with zero volume should not fail: %v", err)
	}
}

func TestExecute_YouTubeAudio_EmptyURL(t *testing.T) {
	appState := &state.AppState{
		EventChannel: make(chan state.AppEvent, 10),
	}
	Init(appState)

	action := models.ActionConfig{
		Type: models.ActionYoutubeAudio,
		URL:  "",
	}

	err := Execute(action)
	if err == nil {
		t.Error("Expected error for empty YouTube URL")
	}
}

func TestExecute_YouTubeAudio_WhitespaceURL(t *testing.T) {
	appState := &state.AppState{
		EventChannel: make(chan state.AppEvent, 10),
	}
	Init(appState)

	action := models.ActionConfig{
		Type: models.ActionYoutubeAudio,
		URL:  "   ",
	}

	err := Execute(action)
	if err == nil {
		t.Error("Expected error for whitespace-only YouTube URL")
	}
}

type mockNotifier struct{}

func (m *mockNotifier) Popup(title, message string) error {
	return nil
}

func (m *mockNotifier) Notify(title, message string) error {
	return nil
}

func (m *mockNotifier) Close() error {
	return nil
}
