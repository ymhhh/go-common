package config

import (
	"os"
	"path/filepath"
	"testing"
)

type demoObj struct {
	N int    `json:"n"`
	S string `json:"s"`
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	return p
}

func TestLoad_JSONC_Include_Ref_Env_GetSet_ToObject(t *testing.T) {
	t.Setenv("ENV", "from-env")

	dir := t.TempDir()

	// included config
	writeFile(t, dir, "inc.yaml", `
a:
  b:
    c: 123
    s: hi
`)

	// main config (JSONC) with comments and include directives
	main := writeFile(t, dir, "main.json", `
// include as directive line
#include inc.yaml
{
  /* include as key too (should be ignored if empty) */
  "#include": [],
  "a": {
    "b": {
      "c": 456, // override included
      "d": "${a.b.c}",
      "e": "${ENV}",
      "mix": "x-${a.b.s}-${ENV}"
    }
  },
  "obj": {
    "n": "${a.b.c}",
    "s": "k"
  }
}
`)

	cfg, err := Load(main)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Get conversions
	if got, _ := cfg.Get("a.b.c").Int(); got != 456 {
		t.Fatalf("a.b.c int: got %d", got)
	}
	if _, ok := cfg.GetOK("a.b.c"); !ok {
		t.Fatalf("GetOK a.b.c: expected ok")
	}
	if _, ok := cfg.GetOK("not.exists"); ok {
		t.Fatalf("GetOK not.exists: expected not ok")
	}
	if got, _ := cfg.Get("a.b.c").Float64(); got != 456 {
		t.Fatalf("a.b.c float64: got %v", got)
	}
	if got, _ := cfg.Get("a.b.d").Int(); got != 456 {
		t.Fatalf("a.b.d ref int: got %d", got)
	}
	if got, _ := cfg.Get("a.b.e").String(); got != "from-env" {
		t.Fatalf("a.b.e env: got %q", got)
	}
	if got, _ := cfg.Get("a.b.mix").String(); got != "x-hi-from-env" {
		t.Fatalf("a.b.mix: got %q", got)
	}

	// Set should create intermediate objects
	if err := cfg.Set("x.y.z", 9); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if got, _ := cfg.Get("x.y.z").Int(); got != 9 {
		t.Fatalf("x.y.z: got %d", got)
	}

	// ToObject
	var obj demoObj
	if err := cfg.ToObject("obj", &obj); err != nil {
		t.Fatalf("ToObject: %v", err)
	}
	if obj.N != 456 || obj.S != "k" {
		t.Fatalf("obj: %+v", obj)
	}
}

func TestIncludeKey_List(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "a.yaml", `
v: 1
`)
	writeFile(t, dir, "b.yaml", `
v: 2
`)
	main := writeFile(t, dir, "main.yaml", `
#include a.yaml
#include b.yaml
v: 3
`)

	cfg, err := Load(main)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got, _ := cfg.Get("v").Int(); got != 3 {
		t.Fatalf("v: got %d", got)
	}
}

func TestLoad_YAML_Ref_Env_ToObject(t *testing.T) {
	t.Setenv("ENV", "yaml-env")

	dir := t.TempDir()

	writeFile(t, dir, "inc.yaml", `
a:
  b:
    c: 7
    s: hello
`)

	main := writeFile(t, dir, "main.yaml", `
#include inc.yaml
a:
  b:
    # override included
    c: 9
    d: ${a.b.c}
    e: ${ENV}
    mix: "p-${a.b.s}-${ENV}"
obj:
  n: ${a.b.c}
  s: kk
`)

	cfg, err := Load(main)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if _, ok := cfg.GetOK("a.b.c"); !ok {
		t.Fatalf("GetOK a.b.c: expected ok")
	}
	if _, ok := cfg.GetOK("not.exists"); ok {
		t.Fatalf("GetOK not.exists: expected not ok")
	}

	if got, _ := cfg.Get("a.b.c").Int(); got != 9 {
		t.Fatalf("a.b.c: got %d", got)
	}
	if got, _ := cfg.Get("a.b.d").Int(); got != 9 {
		t.Fatalf("a.b.d ref: got %d", got)
	}
	if got, _ := cfg.Get("a.b.e").String(); got != "yaml-env" {
		t.Fatalf("a.b.e env: got %q", got)
	}
	if got, _ := cfg.Get("a.b.mix").String(); got != "p-hello-yaml-env" {
		t.Fatalf("a.b.mix: got %q", got)
	}

	// map conversion
	m, err := cfg.Get("a.b").Map()
	if err != nil {
		t.Fatalf("a.b map: %v", err)
	}
	if _, ok := m["c"]; !ok {
		t.Fatalf("a.b map missing c")
	}

	// ToObject
	var obj demoObj
	if err := cfg.ToObject("obj", &obj); err != nil {
		t.Fatalf("ToObject: %v", err)
	}
	if obj.N != 9 || obj.S != "kk" {
		t.Fatalf("obj: %+v", obj)
	}
}

func TestValue_Slice(t *testing.T) {
	sl, err := (Value{v: []any{1, "a"}}).Slice()
	if err != nil {
		t.Fatalf("slice []any: %v", err)
	}
	if len(sl) != 2 {
		t.Fatalf("len: %d", len(sl))
	}

	intSl, err := (Value{v: []int{7, 8}}).Slice()
	if err != nil {
		t.Fatalf("slice []int: %v", err)
	}
	if len(intSl) != 2 {
		t.Fatalf("len: %d", len(intSl))
	}

	jsonSl, err := (Value{v: `[1,2,3]`}).Slice()
	if err != nil {
		t.Fatalf("json string slice: %v", err)
	}
	if len(jsonSl) != 3 {
		t.Fatalf("json len: %d", len(jsonSl))
	}

	if _, err := (Value{v: map[string]any{}}).Slice(); err == nil {
		t.Fatalf("expected error for map")
	}
	if _, err := (Value{v: nil}).Slice(); err == nil {
		t.Fatalf("expected error for nil")
	}
}
