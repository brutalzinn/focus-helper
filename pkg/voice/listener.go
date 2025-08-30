// pkg/voice/listener.go
package voice

import (
	"context"
	"fmt"
	"focus-helper/pkg/actions"
	"focus-helper/pkg/state"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gordonklaus/portaudio"
)

const (
	StateIdle               = "idle"
	StateAwake              = "awake"
	StateProcessing         = "processing"
	StateWaitingForResponse = "waiting_for_response"
)

var wakeTimeout = 30 * time.Second
var lastWakeUp time.Time
var wakeCooldown = 1 * time.Second

const (
	sampleRate        = 16000
	frameMs           = 30
	frameSamples      = sampleRate * frameMs / 1000
	preRollMs         = 300
	preRollFramesMax  = preRollMs / frameMs
	hangoverMs        = 300
	hangoverFrames    = hangoverMs / frameMs
	vadThreshold      = 0.01
	minSpeechMs       = 120
	minSpeechFrames   = minSpeechMs / frameMs
	maxCommandSeconds = 15
	maxChunkSamples   = sampleRate * maxCommandSeconds
)

type Command struct {
	Phrases      []string
	Callback     func(ctx *CommandContext)
	IsActivation bool
}

type CommandContext struct {
	Text      string
	Response  chan string
	RequestID string
}

type audioRingBuffer struct {
	buf       []float32
	headPos   int
	isFull    bool
	frameSize int
}

type Listener struct {
	pendingResponses map[string]chan string
	pendingMu        sync.Mutex
	wakeTimer        *time.Timer
	stream           *portaudio.Stream
	transcriber      *Transcriber
	appState         *state.AppState
	commands         []Command
	inBuffer         []int16
	mainAudioBuffer  []float32
	preRollBuffer    *audioRingBuffer
	stopCh           chan struct{}
	wg               sync.WaitGroup
	closeOnce        sync.Once
	state            string
	stateMu          sync.Mutex
	ReadyCh          chan struct{}
	audioQueue       chan []float32
	speechStartTime  time.Time
	lastSegmentTime  time.Time
}

// newAudioRingBuffer creates a ring for numFrames * samplesPerFrame floats.
func newAudioRingBuffer(numFrames, samplesPerFrame int) *audioRingBuffer {
	return &audioRingBuffer{
		buf:       make([]float32, numFrames*samplesPerFrame),
		frameSize: samplesPerFrame,
	}
}

// PushFrame writes frame into ring buffer handling wrap-around.
func (r *audioRingBuffer) PushFrame(frame []float32) {
	if len(frame) != r.frameSize {
		// defensive: ignore wrong-size frames
		return
	}
	start := r.headPos
	end := start + r.frameSize
	if end <= len(r.buf) {
		copy(r.buf[start:end], frame)
		r.headPos = end % len(r.buf)
	} else {
		first := len(r.buf) - start
		copy(r.buf[start:], frame[:first])
		copy(r.buf[:], frame[first:])
		r.headPos = (start + r.frameSize) % len(r.buf)
	}
	if r.headPos == 0 {
		r.isFull = true
	}
}

// WriteContentsTo writes ring contents (oldest-first) into dst and returns count.
func (r *audioRingBuffer) WriteContentsTo(dst []float32) int {
	if !r.isFull {
		return copy(dst, r.buf[:r.headPos])
	}
	// when full, oldest sample begins at headPos
	n1 := copy(dst, r.buf[r.headPos:])
	n2 := copy(dst[n1:], r.buf[:r.headPos])
	return n1 + n2
}

func NewListener(appState *state.AppState) (*Listener, error) {
	log.Println("Initializing Voice Listener...")

	// Initialize PortAudio
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("portaudio initialize: %w", err)
	}

	in := make([]int16, frameSamples)

	// List devices and prompt user (keeps your original behavior)
	devices, err := portaudio.Devices()
	if err != nil {
		portaudio.Terminate()
		return nil, err
	}

	log.Println("Available input devices:")
	for i, dev := range devices {
		if dev.MaxInputChannels > 0 {
			fmt.Printf("[%d] %s (Input Channels: %d, DefaultSampleRate: %.0f)\n",
				i, dev.Name, dev.MaxInputChannels, dev.DefaultSampleRate)
		}
	}

	var choice int
	for {
		fmt.Print("Select input device by number: ")
		_, err := fmt.Scanf("%d\n", &choice)
		if err != nil || choice < 0 || choice >= len(devices) || devices[choice].MaxInputChannels == 0 {
			fmt.Println("Invalid choice, try again.")
			continue
		}
		break
	}

	selectedDevice := devices[choice]
	log.Printf("Selected device: %s\n", selectedDevice.Name)

	// Force 16kHz mono (Whisper requirement)
	params := portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   selectedDevice,
			Channels: 1,
		},
		SampleRate:      sampleRate,
		FramesPerBuffer: frameSamples,
	}
	params.Output.Channels = 0

	stream, err := portaudio.OpenStream(params, in)
	if err != nil {
		portaudio.Terminate()
		return nil, fmt.Errorf("open stream: %w", err)
	}

	transcriber, err := NewTranscriber(appState.AppConfig.WhisperModelPath)
	if err != nil {
		stream.Close()
		portaudio.Terminate()
		return nil, fmt.Errorf("new transcriber: %w", err)
	}

	listener := &Listener{
		pendingResponses: make(map[string]chan string),
		state:            StateIdle,
		stream:           stream,
		transcriber:      transcriber,
		appState:         appState,
		commands:         []Command{},
		inBuffer:         in,
		mainAudioBuffer:  make([]float32, maxChunkSamples),
		preRollBuffer:    newAudioRingBuffer(preRollFramesMax, frameSamples),
		stopCh:           make(chan struct{}),
		ReadyCh:          make(chan struct{}),
		audioQueue:       make(chan []float32, 2),
	}
	return listener, nil
}

func (l *Listener) ListenContinuously(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Voice command listener started...")

	l.wg.Add(1)
	go l.transcriptionWorker(ctx)

	// Start audio capture loop
	l.wg.Add(1)
	go l.audioCaptureLoop(ctx)

	// Start stream
	if err := l.stream.Start(); err != nil {
		log.Printf("Error starting audio stream: %v", err)
		l.Close()
		return
	}

	close(l.ReadyCh)
	<-ctx.Done()
	log.Println("Shutdown signal received, closing voice listener.")
	l.Close()
}

func (l *Listener) audioCaptureLoop(ctx context.Context) {
	defer l.wg.Done()
	var segmentPos int
	var speechActive bool
	var speechFrames, silenceFrames int
	frameF32 := make([]float32, frameSamples)

	for {
		select {
		case <-ctx.Done():
			return
		case <-l.stopCh:
			return
		default:

			if err := l.stream.Read(); err != nil {
				if err == portaudio.InputOverflowed {
					time.Sleep(5 * time.Millisecond)
					continue
				}
				return
			}

			i16ToF32(l.inBuffer, frameF32)
			l.preRollBuffer.PushFrame(frameF32)
			energy := rmsEnergy(frameF32)
			isSpeech := energy > vadThreshold

			if isSpeech {
				if !speechActive {
					speechActive = true
					segmentPos = l.preRollBuffer.WriteContentsTo(l.mainAudioBuffer)
					l.speechStartTime = time.Now()
				}
				speechFrames++
				silenceFrames = 0
			} else {
				speechFrames = 0
				if speechActive {
					silenceFrames++
				}
			}

			if speechActive {
				if segmentPos+frameSamples <= len(l.mainAudioBuffer) {
					copy(l.mainAudioBuffer[segmentPos:], frameF32)
					segmentPos += frameSamples
				}

				if silenceFrames >= hangoverFrames || segmentPos+frameSamples > len(l.mainAudioBuffer) {
					if segmentPos > 0 {
						segmentCopy := make([]float32, segmentPos)
						copy(segmentCopy, l.mainAudioBuffer[:segmentPos])
						l.lastSegmentTime = time.Now()
						pushAudioSegment(l.audioQueue, segmentCopy)
						log.Printf("Captured speech duration: %v", l.lastSegmentTime.Sub(l.speechStartTime))
					}
					speechActive = false
					speechFrames = 0
					silenceFrames = 0
					segmentPos = 0
				}
			}
		}
	}
}
func (l *Listener) transcriptionWorker(ctx context.Context) {
	defer l.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-l.stopCh:
			return
		case segment := <-l.audioQueue:
			if !l.appState.IsListening {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			text, err := l.transcriber.Transcribe(segment)
			if err != nil {
				log.Printf("Transcription error: %v", err)
				continue
			}
			if text == "" {
				continue
			}

			latency := time.Since(l.lastSegmentTime)
			log.Printf("TRANSCRIBE: %s (latency: %v, segment duration: %v)", text, latency, time.Since(l.speechStartTime))

			l.processCommands(text)
		}
	}
}

func (l *Listener) processCommands(text string) {
	log.Printf("Transcribed speech: '%s'", text)
	normText := normalizeText(text)
	log.Printf("TRANSCRIBE normalize speech: '%s'", normText)
	currentState := l.GetState()
	if currentState == StateWaitingForResponse {
		log.Println("Waiting by user response")
		var delivered bool
		l.pendingMu.Lock()
		for id, respCh := range l.pendingResponses {
			select {
			case respCh <- normText:
				log.Printf("Delivered speech to pending response id=%s", id)
				delivered = true
			default:
				log.Printf("Pending response channel full id=%s", id)
			}
			delete(l.pendingResponses, id)
		}
		l.pendingMu.Unlock()
		if delivered {
			l.SetState(StateIdle)
			return
		}
	}
	for _, wakeCmd := range l.commands {
		if !wakeCmd.IsActivation {
			continue
		}
		for _, phrase := range wakeCmd.Phrases {
			if strings.Contains(normText, phrase) {
				if time.Since(lastWakeUp) < wakeCooldown {
					break
				}
				lastWakeUp = time.Now()
				log.Printf("Wake-up word matched: '%s'", phrase)
				go wakeCmd.Callback(&CommandContext{
					Text:     text,
					Response: make(chan string, 1),
				})
				l.WakeUp()
				return
			}
		}
	}

	if currentState != StateAwake {
		log.Printf("Ignoring commands, listener not awake (state=%s)", currentState)
		return
	}

	if currentState == StateProcessing {
		log.Printf("Ignoring new command detection (state=processing).")
		return
	}

	var matchedCommand *Command
	for i := range l.commands {
		cmd := &l.commands[i]
		if cmd.IsActivation {
			continue
		}
		for _, phrase := range cmd.Phrases {
			if strings.Contains(normText, phrase) {
				matchedCommand = cmd
				log.Printf("Match for phrase: '%s'", phrase)
				break
			}
		}
		if matchedCommand != nil {
			break
		}
	}
	if matchedCommand == nil {
		return
	}
	l.SetState(StateProcessing)
	go l.startCommandWithResponse(matchedCommand, text)
}

func (l *Listener) startCommandWithResponse(cmd *Command, input string) {
	reqID := fmt.Sprintf("%d", time.Now().UnixNano())
	ctx := &CommandContext{
		Text:      input,
		Response:  make(chan string, 1),
		RequestID: reqID,
	}
	l.pendingMu.Lock()
	l.pendingResponses[reqID] = ctx.Response
	l.pendingMu.Unlock()
	l.SetState(StateWaitingForResponse)
	go cmd.Callback(ctx)
}

func (l *Listener) checkForWakeTimeout() {
	currentState := l.GetState()
	if currentState == StateWaitingForResponse {
		log.Printf("Wake timeout check skipped, listener is busy waiting for a response.")
		return
	}
	log.Println("Wake timeout expired, returning to IDLE state.")
	l.SetState(StateIdle)
}

func (l *Listener) SetState(s string) {
	l.stateMu.Lock()
	defer l.stateMu.Unlock()
	if l.state == s {
		return
	}
	log.Printf("Listener state changed: %s → %s", l.state, s)
	l.state = s
}

func (l *Listener) GetState() string {
	l.stateMu.Lock()
	defer l.stateMu.Unlock()
	return l.state
}

func (l *Listener) Close() {
	l.closeOnce.Do(func() {
		close(l.stopCh)
		if l.stream != nil {
			l.stream.Stop()
			l.stream.Close()
		}
		if l.transcriber != nil {
			l.transcriber.Close()
		}
		portaudio.Terminate()
	})
}

func (l *Listener) WakeUp() {
	if err := actions.StopCurrentActions(); err != nil {
		log.Printf("Cant stop actions: %v", err)
	}
	l.SetState(StateAwake)
	log.Println("Listener is now AWAKE, listening for commands.")
	if l.wakeTimer != nil {
		l.wakeTimer.Stop()
	}
	l.wakeTimer = time.AfterFunc(wakeTimeout, func() {
		l.checkForWakeTimeout()
	})
}
