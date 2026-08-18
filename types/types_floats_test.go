package types

import (
	"encoding/json"
	"flag"
	"math"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestFound_FlagValueAndGetter(t *testing.T) {
	var f Found

	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	fs.Var(&f, "fund", "fund")

	if err := fs.Parse([]string{"-fund=12.3"}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if f.String() != "12.30" {
		t.Fatalf("string: got %q", f.String())
	}
	if got := f.Get(); got.(float64) != 12.3 {
		t.Fatalf("get: got %v", got)
	}
}

func TestFound_YAML(t *testing.T) {
	type cfg struct {
		F Found `yaml:"f"`
	}

	var c cfg
	if err := yaml.Unmarshal([]byte("f: 1.234\n"), &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// UnmarshalYAML rounds to 2 decimals via Set(formatFloat(..., 2))
	if c.F.String() != "1.23" {
		t.Fatalf("unmarshal string: got %q", c.F.String())
	}

	out, err := yaml.Marshal(cfg{F: Found(2.5)})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var round cfg
	if err := yaml.Unmarshal(out, &round); err != nil {
		t.Fatalf("roundtrip: %v", err)
	}
	if round.F.String() != "2.50" {
		t.Fatalf("roundtrip: got %q", round.F.String())
	}
}

func TestToFloat64(t *testing.T) {
	if got, err := ToFloat64(nil); err != nil || got != 0 {
		t.Fatalf("nil: got=%v err=%v", got, err)
	}
	if got, err := ToFloat64(float64(1.5)); err != nil || got != 1.5 {
		t.Fatalf("float64: got=%v err=%v", got, err)
	}
	if got, err := ToFloat64(int64(12)); err != nil || got != 12 {
		t.Fatalf("int64: got=%v err=%v", got, err)
	}
	if got, err := ToFloat64(uint32(9)); err != nil || got != 9 {
		t.Fatalf("uint32: got=%v err=%v", got, err)
	}
	if got, err := ToFloat64(uint64(math.MaxUint32)); err != nil || got != float64(math.MaxUint32) {
		t.Fatalf("uint64: got=%v err=%v", got, err)
	}
	if got, err := ToFloat64("2.5"); err != nil || got != 2.5 {
		t.Fatalf("string: got=%v err=%v", got, err)
	}
	if got, err := ToFloat64(json.Number("1e2")); err != nil || got != 100 {
		t.Fatalf("json.Number: got=%v err=%v", got, err)
	}
	if got, err := ToFloat64(Found(3.25)); err != nil || got != 3.25 {
		t.Fatalf("Found: got=%v err=%v", got, err)
	}
	if _, err := ToFloat64(true); err == nil {
		t.Fatalf("expected error for bool")
	} else if !strings.Contains(err.Error(), "cannot convert bool to float64") {
		t.Fatalf("unexpected error: %v", err)
	}
}
