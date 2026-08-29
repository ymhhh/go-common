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
	tests := []struct {
		name    string
		in      any
		want    float64
		wantErr string
	}{
		{name: "nil", in: nil, want: 0},
		{name: "float64", in: float64(1.5), want: 1.5},
		{name: "int64", in: int64(12), want: 12},
		{name: "uint32", in: uint32(9), want: 9},
		{name: "uint64", in: uint64(math.MaxUint32), want: float64(math.MaxUint32)},
		{name: "string", in: "2.5", want: 2.5},
		{name: "json.Number", in: json.Number("1e2"), want: 100},
		{name: "Found", in: Found(3.25), want: 3.25},
		{name: "Fund", in: Fund(3.25), want: 3.25},
		{name: "named int", in: namedInt(9), want: 9},
		{name: "bool", in: true, wantErr: "cannot convert bool to float64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToFloat64(tt.in)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err=%v, want substring %q", err, tt.wantErr)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("got=%v err=%v, want %v", got, err, tt.want)
			}
		})
	}
}
