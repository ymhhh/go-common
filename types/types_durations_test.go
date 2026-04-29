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

	if err := d.Set(""); err != nil {
		t.Fatalf("set empty: %v", err)
	}
	if d.Duration() != 0 {
		t.Fatalf("set empty should reset to 0, got %v", d.Duration())
	}
}

func TestDuration_YAML(t *testing.T) {
	type cfg struct {
		D Duration `yaml:"d"`
	}

	// Unmarshal string duration
	var c cfg
	if err := yaml.Unmarshal([]byte("d: 250ms\n"), &c); err != nil {
		t.Fatalf("unmarshal str: %v", err)
	}
	if c.D.Duration() != 250*time.Millisecond {
		t.Fatalf("unmarshal str: got %v", c.D.Duration())
	}

	// Unmarshal int (nanoseconds)
	var c2 cfg
	if err := yaml.Unmarshal([]byte("d: 1000000\n"), &c2); err != nil {
		t.Fatalf("unmarshal int: %v", err)
	}
	if c2.D.Duration() != time.Millisecond {
		t.Fatalf("unmarshal int: got %v", c2.D.Duration())
	}

	// Marshal should emit string form (e.g. "1.5s")
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

