package types

import (
	"flag"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSecret_StringRedacts(t *testing.T) {
	s := Secret("super-secret")
	if s.String() != Hidden {
		t.Fatalf("String: got %q want %q", s.String(), Hidden)
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

	const block = `
tags:
  - a
  - b
`
	var c1 cfg
	if err := yaml.Unmarshal([]byte(block), &c1); err != nil {
		t.Fatalf("unmarshal block: %v", err)
	}
	if len(c1.Tags) != 2 || c1.Tags[0] != "a" || c1.Tags[1] != "b" {
		t.Fatalf("block: %#v", []string(c1.Tags))
	}

	const flow = `tags: [x, y]`
	var c2 cfg
	if err := yaml.Unmarshal([]byte(flow), &c2); err != nil {
		t.Fatalf("unmarshal flow: %v", err)
	}
	if len(c2.Tags) != 2 || c2.Tags[0] != "x" || c2.Tags[1] != "y" {
		t.Fatalf("flow: %#v", []string(c2.Tags))
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
