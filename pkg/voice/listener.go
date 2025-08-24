// pkg/voice/listener.go
package voice

import (
	"context"
	"fmt"
	"focus-helper/pkg/models"
	"focus-helper/pkg/state"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gordonklaus/portaudio"
)

// --- Constants for Audio Processing ---
const (
	sampleRate   = 16000 // Sample rate expected by Whisper.
	frameMs      = 30    // Duration of each audio frame in milliseconds.
	frameSamples = sampleRate * frameMs / 1000

	preRollMs        = 300 // Keep 300ms of audio before speech starts.
	preRollFramesMax = preRollMs / frameMs

	hangoverMs     = 300 // Wait 300ms after speech ends before processing.
	hangoverFrames = hangoverMs / frameMs

	vadThreshold    = 0.01 // Voice Activity Detection (VAD) threshold. Tune this for your mic.
	minSpeechMs     = 120  // Minimum duration of speech to be considered active.
	minSpeechFrames = minSpeechMs / frameMs

	maxSpeechSeconds  = 15                                  // Max duration of a single speech segment.
	maxSegmentFrames  = (maxSpeechSeconds * 1000) / frameMs // Max frames in the main buffer.
	maxSegmentSamples = maxSegmentFrames * frameSamples     // Max samples in the main buffer.

	transcriptionWorkers = 1 // Number of parallel transcription workers.
)

var wakeTimeout = 15 * time.Second

type Command struct {
	Phrases      []string
	Callback     func(transcribedText string)
	IsActivation bool
}

type audioRingBuffer struct {
	buf     []float32
	headPos int
	isFull  bool
}

type Listener struct {
	stream          *portaudio.Stream
	transcriber     *Transcriber
	appState        *state.AppState
	commands        []Command
	inBuffer        []int16
	frameBuffer     []float32
	mainAudioBuffer []float32
	preRollBuffer   *audioRingBuffer
	stopCh          chan struct{}
	wg              sync.WaitGroup
	transcriptionCh chan []float32
	closeOnce       sync.Once
	state           string
	stateMu         sync.Mutex
}

// DISCLAIMER: SOMETHING CAN THORW SEGMENTION FAULT. wtf many GOROUTINES FOR EVERYTHING.
///NO ONE UNIT TEST. ARE YOU JOKING ME???? @brutalzinn
// SOMEPOINTERS IS WRONG BUT WE ARE APPLYING GO HORSE NOW

func newAudioRingBuffer(numFrames, samplesPerFrame int) *audioRingBuffer {
	return &audioRingBuffer{
		buf: make([]float32, numFrames*samplesPerFrame),
	}
}

func (r *audioRingBuffer) PushFrame(frame []float32) {
	frameSize := len(frame)
	copy(r.buf[r.headPos:], frame)
	r.headPos += frameSize
	if r.headPos >= len(r.buf) {
		r.headPos = 0
		r.isFull = true
	}
}

func (r *audioRingBuffer) WriteContentsTo(dst []float32) int {
	if !r.isFull {
		return copy(dst, r.buf[:r.headPos])
	}

	copied := copy(dst, r.buf[r.headPos:])
	copied += copy(dst[copied:], r.buf[:r.headPos])
	return copied
}

func NewListener(appState *state.AppState) (*Listener, error) {
	log.Println("Initializing Voice Listener...")

	in := make([]int16, frameSamples)

	devices, err := portaudio.Devices()
	if err != nil {
		return nil, err
	}

	log.Println("Available input devices:")
	for i, dev := range devices {
		if dev.MaxInputChannels > 0 {
			fmt.Printf("[%d] %s (Input Channels: %d)\n", i, dev.Name, dev.MaxInputChannels)
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

	params := portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   selectedDevice,
			Channels: 1,
		},
		SampleRate:      sampleRate,
		FramesPerBuffer: len(in),
	}
	params.Output.Channels = 0
	stream, err := portaudio.OpenStream(params, in)
	if err != nil {
		return nil, err
	}

	transcriber, err := NewTranscriber(appState.AppConfig.WhisperModelPath)
	if err != nil {
		stream.Close()
		return nil, err
	}
	listener := &Listener{
		state:           "idle",
		stream:          stream,
		appState:        appState,
		transcriber:     transcriber,
		commands:        make([]Command, 0),
		inBuffer:        in,
		frameBuffer:     make([]float32, frameSamples),
		mainAudioBuffer: make([]float32, maxSegmentSamples),
		preRollBuffer:   newAudioRingBuffer(preRollFramesMax, frameSamples),
		stopCh:          make(chan struct{}),
		transcriptionCh: make(chan []float32, transcriptionWorkers),
	}

	return listener, nil
}

func (l *Listener) ListenContinuously(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Voice command listener started...")
	l.wg.Add(transcriptionWorkers)
	for i := 0; i < transcriptionWorkers; i++ {
		go l.transcriptionWorker(ctx)
	}
	l.wg.Add(1)
	go l.audioCaptureLoop(ctx)
	if err := l.stream.Start(); err != nil {
		log.Printf("Error starting audio stream: %v", err)
		l.Close()
	}
	<-ctx.Done()
	log.Println("Shutdown signal received, closing voice listener.")
	l.Close()
}

func (l *Listener) audioCaptureLoop(ctx context.Context) {
	defer l.wg.Done()
	var segmentPos int
	var speechActive bool
	var speechFrames, silenceFrames int
	for {
		select {
		case <-ctx.Done():
			log.Println("Audio capture loop stopping due to context cancellation.")
			return
		case <-l.stopCh:
			log.Println("Audio capture loop stopped.")
			return
		default:
			if err := l.stream.Read(); err != nil {
				log.Printf("Error reading from audio stream: %v", err)
				return
			}
			i16ToF32(l.inBuffer, l.frameBuffer)
			l.preRollBuffer.PushFrame(l.frameBuffer)
			energy := rmsEnergy(l.frameBuffer)
			isSpeech := energy > vadThreshold
			if isSpeech {
				speechFrames++
				silenceFrames = 0
				if !speechActive && speechFrames >= minSpeechFrames {
					speechActive = true
					segmentPos = l.preRollBuffer.WriteContentsTo(l.mainAudioBuffer)
				}
			} else {
				speechFrames = 0
				if speechActive {
					silenceFrames++
				}
			}
			if speechActive {
				if segmentPos+frameSamples <= len(l.mainAudioBuffer) {
					copy(l.mainAudioBuffer[segmentPos:], l.frameBuffer)
					segmentPos += frameSamples
				}
				if silenceFrames >= hangoverFrames || segmentPos+frameSamples > len(l.mainAudioBuffer) {
					segmentCopy := make([]float32, segmentPos)
					copy(segmentCopy, l.mainAudioBuffer[:segmentPos])
					l.transcriptionCh <- segmentCopy
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
			log.Println("Transcription worker stopping due to context cancellation.")
			return
		case <-l.stopCh:
			log.Println("Transcription worker stopped.")
			return
		case audioData := <-l.transcriptionCh:
			text, err := l.transcriber.Transcribe(audioData)
			if err != nil {
				log.Printf("Transcription error: %v", err)
				continue
			}
			if text == "" {
				continue
			}

			l.processCommands(text)
		}
	}
}
func (l *Listener) processCommands(text string) {
	log.Printf("Transcribed speech: '%s'", text)
	normText := normalizeText(text)
	log.Printf("Normalized speech: '%s'", normText)
	for _, command := range l.commands {
		var matchedPhrase string
		for _, phrase := range command.Phrases {
			if strings.Contains(normText, phrase) {
				matchedPhrase = phrase
				break
			}
		}
		if matchedPhrase == "" {
			continue
		}
		if command.IsActivation {
			log.Printf("Activation phrase matched: '%s'. Executing callback.", matchedPhrase)
			go command.Callback(text)
			return
		}
		currentState := l.GetState()
		if currentState == "listening" {
			if l.GetState() == "action_running" {
				log.Printf("Action in progress, ignoring command: '%s'", matchedPhrase)
				return
			}
			log.Printf("Command matched: '%s'. Executing callback.", matchedPhrase)
			l.SetState("action_running")
			go func() {
				defer l.SetState("listening")
				command.Callback(text)
			}()
			return
		}
		log.Printf("Command phrase '%s' detected but ignored (current state: %s).", matchedPhrase, currentState)
		return
	}
}

func (l *Listener) SetState(s string) {
	l.stateMu.Lock()
	defer l.stateMu.Unlock()
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
		log.Println("Closing Voice Listener resources...")
		if l.stream != nil {
			l.stream.Stop()
			l.stream.Close()
		}
		close(l.transcriptionCh)
		l.wg.Wait()
		if l.transcriber != nil {
			l.transcriber.Close()
		}
		log.Println("Voice Listener closed.")
	})
}

func (l *Listener) AppConfig() *models.Config {
	return l.appState.AppConfig
}
