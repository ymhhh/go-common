package types

import (
	"flag"
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

