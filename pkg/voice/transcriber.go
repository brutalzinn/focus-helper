package voice

import (
	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

type Transcriber struct {
	model whisper.Model
}

func NewTranscriber(modelPath string) (*Transcriber, error) {
	model, err := whisper.New(modelPath)
	if err != nil {
		return nil, err
	}
	return &Transcriber{model: model}, nil
}

func (t *Transcriber) Transcribe(audio []float32) (string, error) {

	ctxt, err := t.model.NewContext()
	if err != nil {
		return "", err
	}
	ctxt.SetLanguage("pt")
	// ctxt.SetTranslate(true)

	if err := ctxt.Process(audio, nil, nil, nil); err != nil {
		return "", err
	}

	var fullText string
	for {
		seg, err := ctxt.NextSegment()
		if err != nil {
			break
		}
		fullText += seg.Text
	}

	return fullText, nil
}

func (t *Transcriber) Close() {
	if t.model != nil {
		t.model.Close()
	}
}
