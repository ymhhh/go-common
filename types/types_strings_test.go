package types

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSecret_StringRedacts(t *testing.T) {
	s := Secret("super-secret")
	if s.String() != Hidden {
		t.Fatalf("String: got %q want %q", s.String(), Hidden)
	}
}

func TestSecret_GoStringRedacts(t *testing.T) {
	s := Secret("super-secret")
	got := fmt.Sprintf("%#v", s)
	if got != `types.Secret("<hidden>")` {
		t.Fatalf("GoString: got %q", got)
	}
	if got == `types.Secret("super-secret")` {
		t.Fatal("GoString exposed the secret value")
	}

	container := fmt.Sprintf("%#v", struct {
		Token Secret
	}{Token: s})
	if strings.Contains(container, "super-secret") {
		t.Fatal("container GoString exposed the secret value")
	}
}

func TestSecret_FlagValue(t *testing.T) {
	var s Secret

	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	fs.Var(&s, "token", "token")

	if err := fs.Parse([]string{"-token=abc123"}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if string(s) != "abc123" {
		t.Fatalf("underlying: got %q", string(s))
	}
	if s.String() != Hidden {
		t.Fatalf("String should stay redacted: got %q", s.String())
	}
}

func TestSecret_YAML(t *testing.T) {
	type cfg struct {
		S Secret `yaml:"s"`
	}

	in := cfg{S: Secret("xyzzy")}
	out, err := yaml.Marshal(&in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := yaml.Unmarshal(out, &m); err != nil {
		t.Fatalf("unmarshal marshal output: %v", err)
	}
	if got := m["s"]; got != Hidden {
		t.Fatalf("marshal emitted field s=%v want %q", got, Hidden)
	}

	const yamlIn = "s: real-secret-value\n"
	var round cfg
	if err := yaml.Unmarshal([]byte(yamlIn), &round); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if string(round.S) != "real-secret-value" {
		t.Fatalf("unmarshal value: got %q", string(round.S))
	}
	if round.S.String() != Hidden {
		t.Fatalf("String after unmarshal: got %q", round.S.String())
	}

	var empty cfg
	if err := yaml.Unmarshal([]byte("s:\n"), &empty); err != nil {
		t.Fatalf("empty scalar: %v", err)
	}
	if string(empty.S) != "" {
		t.Fatalf("empty scalar: got %q", string(empty.S))
	}

	var seq cfg
	if err := yaml.Unmarshal([]byte("s:\n  - not-a-secret\n"), &seq); err == nil {
		t.Fatalf("sequence should not unmarshal as Secret")
	}
}

func TestSecret_JSON(t *testing.T) {
	type cfg struct {
		S Secret `json:"s"`
	}

	out, err := json.Marshal(cfg{S: Secret("real-secret-value")})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]string
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("unmarshal marshal output: %v", err)
	}
	if got := m["s"]; got != Hidden {
		t.Fatalf("marshal emitted field s=%v want %q", got, Hidden)
	}

	var round cfg
	if err := json.Unmarshal([]byte(`{"s":"real-secret-value"}`), &round); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if string(round.S) != "real-secret-value" {
		t.Fatalf("unmarshal value: got %q", string(round.S))
	}
}

func TestStrings_FlagValue_Append(t *testing.T) {
	var xs Strings

	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	fs.Var(&xs, "tag", "tag")

	if err := fs.Parse([]string{"-tag=a", "-tag=b"}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(xs) != 2 || xs[0] != "a" || xs[1] != "b" {
		t.Fatalf("got %#v", []string(xs))
	}
}

func TestStrings_YAML_SequenceAndFlow(t *testing.T) {
	type cfg struct {
		Tags Strings `yaml:"tags"`
	}

	tests := []struct {
		name string
		in   string
		want []string
	}{
		{name: "block", in: "tags:\n  - a\n  - b\n", want: []string{"a", "b"}},
		{name: "flow", in: `tags: [x, y]`, want: []string{"x", "y"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c cfg
			if err := yaml.Unmarshal([]byte(tt.in), &c); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if len(c.Tags) != len(tt.want) {
				t.Fatalf("got %#v, want %#v", []string(c.Tags), tt.want)
			}
			for i, w := range tt.want {
				if c.Tags[i] != w {
					t.Fatalf("got %#v, want %#v", []string(c.Tags), tt.want)
				}
			}
		})
	}

	out, err := yaml.Marshal(&cfg{Tags: Strings{"p", "q"}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := yaml.Unmarshal(out, &m); err != nil {
		t.Fatalf("roundtrip parse: %v", err)
	}
	tags, ok := m["tags"].([]any)
	if !ok || len(tags) != 2 {
		t.Fatalf("marshal tags: %#v", m["tags"])
	}
}
