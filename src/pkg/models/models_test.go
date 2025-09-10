package models

import (
	"encoding/json"
	"strings"
	"testing"
)

type Ex struct {
	S Duration
	I Duration
}

func Test_If_Decode_Duration(t *testing.T) {
	var ex Ex
	in := strings.NewReader(`{"S": "15s350ms", "I": 400000}`)
	err := json.NewDecoder(in).Decode(&ex)
	if err != nil {
		t.Log(err)
	}
	t.Log("Decoded:", ex)

	out, err := json.Marshal(ex)
	if err != nil {
		panic(err)
	}

	t.Log("Encoded:", string(out))
}
