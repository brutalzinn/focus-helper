package voice

import (
	"focus-helper/pkg/models"
	"focus-helper/pkg/state"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gordonklaus/portaudio"
)

const (
	sampleRate       = 16000
	frameMs          = 30
	frameSamples     = sampleRate * frameMs / 1000
	preRollMs        = 300
	hangoverMs       = 300
	preRollFramesMax = preRollMs / frameMs
	hangoverFrames   = hangoverMs / frameMs
	vadThreshold     = 0.01 // You can tune this threshold
	minSpeechMs      = 120
	minSpeechFrames  = minSpeechMs / frameMs
)

// Ring is a generic circular buffer for pre-roll audio.
type Ring[T any] struct {
	buf  []T
	head int
	fill int
}

type Listener struct {
	Stream      *portaudio.Stream
	inBuffer    []int16
	Transcriber *Transcriber
	AppState    *state.AppState
	AppConfig   *models.Config
}

func NewRing[T any](cap int) *Ring[T] { return &Ring[T]{buf: make([]T, cap)} }

func (r *Ring[T]) Push(x T) {
	if len(r.buf) == 0 {
		return
	}
	r.buf[r.head] = x
	r.head = (r.head + 1) % len(r.buf)
	if r.fill < len(r.buf) {
		r.fill++
	}
}
func (r *Ring[T]) Snapshot() []T {
	out := make([]T, r.fill)
	start := (r.head - r.fill + len(r.buf)) % len(r.buf)
	for i := 0; i < r.fill; i++ {
		out[i] = r.buf[(start+i)%len(r.buf)]
	}
	return out
}

func NewListener(cfg *models.Config, appState *state.AppState) (*Listener, error) {
	log.Println("Initializing Voice Listener...")

	if err := portaudio.Initialize(); err != nil {
		return nil, err
	}

	in := make([]int16, frameSamples)
	stream, err := portaudio.OpenDefaultStream(1, 0, float64(sampleRate), len(in), in)
	if err != nil {
		portaudio.Terminate()
		return nil, err
	}

	if err := stream.Start(); err != nil {
		portaudio.Terminate()
		return nil, err
	}

	transcriber, err := NewTranscriber(cfg.WhisperModelPath)
	if err != nil {
		portaudio.Terminate()
		return nil, err
	}

	return &Listener{
		AppState:    appState,
		AppConfig:   cfg,
		Stream:      stream,
		inBuffer:    in,
		Transcriber: transcriber,
	}, nil
}
func (l *Listener) ListenContinuously(callback func(text string)) {
	preRoll := NewRing[[]float32](preRollFramesMax)
	var segment [][]float32
	var speechActive bool
	var speechCount, silenceCount int

	f32 := make([]float32, frameSamples)

	log.Println("Listening continuously...")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	ticker := time.NewTicker(time.Duration(frameMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-sig:
			log.Println("Interrupt signal received, stopping listener.")
			return
		case <-ticker.C:
			if err := l.Stream.Read(); err != nil {
				log.Printf("Error reading from audio stream: %v", err)
				continue
			}
			i16ToF32(l.inBuffer, f32)
			frameCopy := make([]float32, len(f32))
			copy(frameCopy, f32)
			preRoll.Push(frameCopy)

			energy := rmsEnergy(f32)
			isSpeech := energy > vadThreshold

			if isSpeech {
				speechCount++
				if !speechActive && speechCount >= minSpeechFrames {
					speechActive = true
					silenceCount = 0
					segment = append([][]float32(nil), preRoll.Snapshot()...)
					segment = append(segment, frameCopy)
				} else if speechActive {
					segment = append(segment, frameCopy)
					silenceCount = 0
				}
			} else if speechActive {
				silenceCount++
				if silenceCount >= hangoverFrames {
					audio := flatten(segment)
					go func(audioData []float32) {
						text, err := l.Transcriber.Transcribe(audioData)
						if err == nil && text != "" {
							callback(text)
						}
					}(audio)
					speechActive = false
					silenceCount = 0
					speechCount = 0
					segment = nil
				}
			} else {
				speechCount = 0
			}
		}
	}
}

func (l *Listener) Close() {
	log.Println("Closing Voice Listener...")
	if l.Transcriber != nil {
		l.Transcriber.Close()
	}
	if l.Stream != nil {
		l.Stream.Stop()
		l.Stream.Close()
	}
	portaudio.Terminate()
}
