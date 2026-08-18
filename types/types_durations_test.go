package types

import (
	"flag"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestDuration_FlagValue(t *testing.T) {
	var d Duration

	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	fs.Var(&d, "timeout", "timeout")

	if err := fs.Parse([]string{"-timeout=150ms"}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got := d.Duration(); got != 150*time.Millisecond {
		t.Fatalf("duration: got %v", got)
	}
	if d.String() != "150ms" {
		t.Fatalf("string: got %q", d.String())
	}

	sets := []struct {
		in   string
		want time.Duration
	}{
		{in: "", want: 0},
		{in: "1d", want: 24 * time.Hour},
		{in: "0d", want: 0},
	}
	for _, tt := range sets {
		t.Run("set "+tt.in, func(t *testing.T) {
			if err := d.Set(tt.in); err != nil {
				t.Fatalf("set %q: %v", tt.in, err)
			}
			if d.Duration() != tt.want {
				t.Fatalf("got %v, want %v", d.Duration(), tt.want)
			}
		})
	}
}

func TestDuration_YAML(t *testing.T) {
	type cfg struct {
		D Duration `yaml:"d"`
	}

	tests := []struct {
		name    string
		in      string
		want    time.Duration
		wantErr bool
	}{
		{name: "string", in: "d: 250ms\n", want: 250 * time.Millisecond},
		{name: "fractional", in: "d: 1.5s\n", want: 1500 * time.Millisecond},
		{name: "compound", in: "d: 1h30m\n", want: 90 * time.Minute},
		{name: "int nanos", in: "d: 1000000\n", want: time.Millisecond},
		{name: "invalid", in: "d: definitely-not-a-duration\n", wantErr: true},
		{name: "zero 0", in: "d: 0\n", want: 0},
		{name: "zero 0s", in: "d: 0s\n", want: 0},
		{name: "zero 0d", in: "d: 0d\n", want: 0},
		{name: "zero 0w", in: "d: 0w\n", want: 0},
		{name: "zero 0y", in: "d: 0y\n", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c cfg
			err := yaml.Unmarshal([]byte(tt.in), &c)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if c.D.Duration() != tt.want {
				t.Fatalf("got %v, want %v", c.D.Duration(), tt.want)
			}
		})
	}

	out, err := yaml.Marshal(cfg{D: Duration(1500 * time.Millisecond)})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var round cfg
	if err := yaml.Unmarshal(out, &round); err != nil {
		t.Fatalf("roundtrip: %v", err)
	}
	if round.D.Duration() != 1500*time.Millisecond {
		t.Fatalf("roundtrip: got %v", round.D.Duration())
	}
}
