package config

import (
	"encoding"
	"testing"
)

type textConfig string

var _ encoding.TextUnmarshaler = (*textConfig)(nil)

func (t *textConfig) UnmarshalText(text []byte) error {
	*t = textConfig(string(text))
	return nil
}

type typeOnly struct {
	Type string `json:"type" yaml:"type"`
}

func TestDecodeToObject_string_TextUnmarshaler(t *testing.T) {
	var out textConfig
	if err := decodeToObject("prometheus", &out); err != nil {
		t.Fatalf("err: %v", err)
	}
	if out != "prometheus" {
		t.Fatalf("out: %q", out)
	}
}

func TestDecodeToObject_string_structTypeField(t *testing.T) {
	var out typeOnly
	if err := decodeToObject("prometheus", &out); err != nil {
		t.Fatalf("err: %v", err)
	}
	if out.Type != "prometheus" {
		t.Fatalf("Type: %q", out.Type)
	}
}

