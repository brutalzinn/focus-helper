// pkg/voice/listener.go
package voice

import (
	"fmt"
	"focus-helper/pkg/audio"
	"focus-helper/pkg/models"
	"focus-helper/pkg/state"
	"log"
	"strings"
	"sync"

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

type Command struct {
	Phrase   string
	Callback func(transcribedText string)
}

type audioRingBuffer struct {
	buf     []float32
	headPos int
	isFull  bool
}

type Listener struct {
	stream      *portaudio.Stream
	transcriber *Transcriber
	appState    *state.AppState
	appConfig   *models.Config
	commands    map[string]Command

	inBuffer        []int16
	frameBuffer     []float32
	mainAudioBuffer []float32
	preRollBuffer   *audioRingBuffer

	stopCh          chan struct{}
	wg              sync.WaitGroup
	transcriptionCh chan []float32
	closeOnce       sync.Once
}

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

func NewListener(cfg *models.Config, appState *state.AppState) (*Listener, error) {
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

	transcriber, err := NewTranscriber(cfg.WhisperModelPath)
	if err != nil {
		stream.Close()
		return nil, err
	}
	listener := &Listener{
		appState:        appState,
		appConfig:       cfg,
		stream:          stream,
		transcriber:     transcriber,
		commands:        make(map[string]Command),
		inBuffer:        in,
		frameBuffer:     make([]float32, frameSamples),
		mainAudioBuffer: make([]float32, maxSegmentSamples),
		preRollBuffer:   newAudioRingBuffer(preRollFramesMax, frameSamples),
		stopCh:          make(chan struct{}),
		transcriptionCh: make(chan []float32, transcriptionWorkers),
	}

	return listener, nil
}

func (l *Listener) AppConfig() *models.Config {
	return l.appConfig
}

func (l *Listener) ListenContinuously() {
	log.Println("Voice command listener started...")
	l.wg.Add(transcriptionWorkers)
	for i := 0; i < transcriptionWorkers; i++ {
		go l.transcriptionWorker()
	}
	l.wg.Add(1)
	go l.audioCaptureLoop()
	if err := l.stream.Start(); err != nil {
		log.Printf("Error starting audio stream: %v", err)
		l.Close()
	}
}

func (l *Listener) audioCaptureLoop() {
	defer l.wg.Done()
	var segmentPos int
	var speechActive bool
	var speechFrames, silenceFrames int
	for {
		select {
		case <-l.stopCh:
			log.Println("Audio capture loop stopped.")
			return
		default:
			if err := l.stream.Read(); err != nil {
				log.Printf("Error reading from audio stream: %v", err)
				continue
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

func (l *Listener) transcriptionWorker() {
	defer l.wg.Done()
	for {
		select {
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
	release := audio.RequestAccess()
	defer release()
	log.Printf("Transcribed speech: '%s'", text)
	lowerText := strings.ToLower(text)
	for phrase, command := range l.commands {
		if strings.Contains(lowerText, phrase) {
			log.Printf("Voice command matched for phrase: '%s'", command.Phrase)
			go command.Callback(text)
			return
		}
	}
}

func (l *Listener) Close() {
	l.closeOnce.Do(func() {
		log.Println("Closing Voice Listener...")
		close(l.stopCh)
		if l.stream != nil {
			l.stream.Stop()
			l.stream.Close()
		}
		l.wg.Wait()
		if l.transcriber != nil {
			l.transcriber.Close()
		}
		log.Println("Voice Listener closed.")
	})
}
