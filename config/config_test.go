package config

import (
	"os"
	"path/filepath"
	"strings"
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

func TestLoad_JSONC_Include_Ref_Env_GetSet_Object(t *testing.T) {
	t.Setenv("ENV", "from-env")
	cfg := mustLoadJSONCIncludeFixture(t)

	t.Run("override and refs", func(t *testing.T) {
		assertJSONCOverrideAndRefs(t, cfg)
	})
	t.Run("set", func(t *testing.T) {
		assertJSONCSet(t, cfg)
	})
	t.Run("object", func(t *testing.T) {
		assertJSONCObject(t, cfg)
	})
}

func mustLoadJSONCIncludeFixture(t *testing.T) Config {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, dir, "inc.yaml", `
a:
  b:
    c: 123
    s: hi
`)
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
	return cfg
}

func assertJSONCOverrideAndRefs(t *testing.T, cfg Config) {
	t.Helper()
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
}

func assertJSONCSet(t *testing.T, cfg Config) {
	t.Helper()
	if err := cfg.Set("x.y.z", 9); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if got, _ := cfg.Get("x.y.z").Int(); got != 9 {
		t.Fatalf("x.y.z: got %d", got)
	}
	if err := cfg.Set("a.b.c.leaf", 10); err == nil {
		t.Fatalf("Set through non-object: expected error")
	}
	if got, _ := cfg.Get("a.b.c").Int(); got != 456 {
		t.Fatalf("a.b.c should remain unchanged after failed Set: got %d", got)
	}
}

func assertJSONCObject(t *testing.T, cfg Config) {
	t.Helper()
	var obj demoObj
	if err := cfg.Object(&obj, WithObjectPath("obj")); err != nil {
		t.Fatalf("Object: %v", err)
	}
	if obj.N != 456 || obj.S != "k" {
		t.Fatalf("obj: %+v", obj)
	}
}

func TestLoad_RefPrefersConfigPathOverEnvironment(t *testing.T) {
	t.Setenv("secrets.db.password", "from-env")

	dir := t.TempDir()
	main := writeFile(t, dir, "main.yaml", `
secrets:
  db:
    password: from-config
database:
  password: ${secrets.db.password}
`)

	cfg, err := Load(main)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := cfg.GetString("database.password"); got != "from-config" {
		t.Fatalf("database.password: got %q", got)
	}
}

func TestLoad_YAML_IndentedIncludeTextPreserved(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "override.yaml", `
auth:
  enabled: false
`)
	main := writeFile(t, dir, "main.yaml", `
auth:
  enabled: true
notes: |
  Deployment note:
  #include override.yaml
`)

	cfg, err := Load(main)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.GetBoolean("auth.enabled") {
		t.Fatalf("indented include text should not load override.yaml")
	}
	if got := cfg.GetString("notes"); !strings.Contains(got, "#include override.yaml") {
		t.Fatalf("notes should preserve indented include text, got %q", got)
	}
}

func TestLoad_JSONC_CommentedIncludeIgnored(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "override.yaml", `
auth:
  enabled: false
  admin: true
`)
	main := writeFile(t, dir, "main.json", `
/*
#include override.yaml
*/
{
  "auth": {
    "enabled": true
  }
}
`)

	cfg, err := Load(main)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.GetBoolean("auth.enabled") {
		t.Fatalf("commented include should not load override.yaml")
	}
	if _, ok := cfg.GetOK("auth.admin"); ok {
		t.Fatalf("commented include should not merge auth.admin")
	}
}

func TestLoad_JSONC_CommentStrippingDoesNotCreateInclude(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "override.yaml", `
auth:
  admin: true
`)
	main := writeFile(t, dir, "main.json", `
/**/#include override.yaml
{
  "auth": {}
}
`)

	if _, err := Load(main); err == nil {
		t.Fatalf("expected invalid JSONC instead of manufacturing an include directive")
	} else if strings.Contains(err.Error(), "override.yaml") {
		t.Fatalf("error should come from parsing main config, not loading override: %v", err)
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

func TestLoad_JSONC_UnterminatedBlockComment(t *testing.T) {
	dir := t.TempDir()
	main := writeFile(t, dir, "bad.json", `{
  "safe": true
  /* missing close
`)

	if _, err := Load(main); err == nil {
		t.Fatalf("expected unterminated block comment error")
	}
}

func TestLoad_YAML_Ref_Env_Object(t *testing.T) {
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

	// Object
	var obj demoObj
	if err := cfg.Object(&obj, WithObjectPath("obj")); err != nil {
		t.Fatalf("Object: %v", err)
	}
	if obj.N != 9 || obj.S != "kk" {
		t.Fatalf("obj: %+v", obj)
	}
}

func TestLoad_YAML_MultipleDocumentsMerged(t *testing.T) {
	dir := t.TempDir()
	main := writeFile(t, dir, "main.yaml", `
a:
  b:
    keep: base
    override: old
---
a:
  b:
    override: new
    added: value
`)

	cfg, err := Load(main)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := cfg.GetString("a.b.keep"); got != "base" {
		t.Fatalf("a.b.keep: got %q", got)
	}
	if got := cfg.GetString("a.b.override"); got != "new" {
		t.Fatalf("a.b.override: got %q", got)
	}
	if got := cfg.GetString("a.b.added"); got != "value" {
		t.Fatalf("a.b.added: got %q", got)
	}
}

func TestLoad_JSONRejectsTrailingTopLevelValue(t *testing.T) {
	dir := t.TempDir()
	main := writeFile(t, dir, "main.json", `
{"auth": {"enabled": false}}
{"auth": {"enabled": true}}
`)

	if _, err := Load(main); err == nil {
		t.Fatalf("expected trailing JSON value to be rejected")
	}
}

func TestLoad_YAML_CompositeReferenceIsDeepCopied(t *testing.T) {
	dir := t.TempDir()
	main := writeFile(t, dir, "main.yaml", `
base:
  host: db.internal
  limits:
    retries: 3
derived: ${base}
items:
  - alpha
  - beta
itemsCopy: ${items}
`)

	cfg, err := Load(main)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if err := cfg.Set("derived.host", "db.override"); err != nil {
		t.Fatalf("Set derived.host: %v", err)
	}
	if got := cfg.GetString("base.host"); got != "db.internal" {
		t.Fatalf("base.host should not change after mutating derived: got %q", got)
	}

	derived := cfg.GetMap("derived")
	derivedLimits, ok := derived["limits"].(map[string]any)
	if !ok {
		t.Fatalf("derived.limits: got %T", derived["limits"])
	}
	derivedLimits["retries"] = 5
	if got := cfg.GetInt("base.limits.retries"); got != 3 {
		t.Fatalf("base.limits.retries should not change after mutating derived: got %d", got)
	}

	itemsCopy := cfg.GetList("itemsCopy")
	itemsCopy[0] = "changed"
	if got := cfg.GetStringList("items")[0]; got != "alpha" {
		t.Fatalf("items should not change after mutating itemsCopy: got %q", got)
	}
}

func TestResolve_CompositeReferenceCopiesProgrammaticMutableValues(t *testing.T) {
	opts := Options{
		"typedMap":      map[string]string{"host": "db.internal"},
		"typedMapCopy":  "${typedMap}",
		"typedList":     []string{"alpha", "beta"},
		"typedListCopy": "${typedList}",
	}
	cfg := (&opts).ToConfig()

	if err := cfg.Resolve(); err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	typedMapCopy, ok := cfg.Get("typedMapCopy").Any().(map[string]string)
	if !ok {
		t.Fatalf("typedMapCopy: got %T", cfg.Get("typedMapCopy").Any())
	}
	typedMapCopy["host"] = "db.override"
	if got := cfg.Get("typedMap").Any().(map[string]string)["host"]; got != "db.internal" {
		t.Fatalf("typedMap should not change after mutating typedMapCopy: got %q", got)
	}

	typedListCopy, ok := cfg.Get("typedListCopy").Any().([]string)
	if !ok {
		t.Fatalf("typedListCopy: got %T", cfg.Get("typedListCopy").Any())
	}
	typedListCopy[0] = "changed"
	if got := cfg.Get("typedList").Any().([]string)[0]; got != "alpha" {
		t.Fatalf("typedList should not change after mutating typedListCopy: got %q", got)
	}
}

func TestValue_Slice(t *testing.T) {
	tests := []struct {
		name    string
		in      any
		wantLen int
		wantErr bool
	}{
		{name: "[]any", in: []any{1, "a"}, wantLen: 2},
		{name: "[]int", in: []int{7, 8}, wantLen: 2},
		{name: "json string", in: `[1,2,3]`, wantLen: 3},
		{name: "map", in: map[string]any{}, wantErr: true},
		{name: "nil", in: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (Value{v: tt.in}).Slice()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil || len(got) != tt.wantLen {
				t.Fatalf("len=%d err=%v, want %d", len(got), err, tt.wantLen)
			}
		})
	}
}

func TestResolve_ReferenceCycle(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.yaml", `
x: ${y}
y: ${x}
`)
	_, err := Load(filepath.Join(dir, "a.yaml"))
	if err == nil {
		t.Fatalf("expected cycle error")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("error should mention cycle: %v", err)
	}
}

func TestResolve_UndefinedReferenceReturnsError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.yaml", `
host: ${UNDEFINED_VAR_XYZ}
`)
	_, err := Load(filepath.Join(dir, "a.yaml"))
	if err == nil {
		t.Fatalf("expected error for undefined reference")
	}
	if !strings.Contains(err.Error(), "unresolved") {
		t.Fatalf("error should mention unresolved: %v", err)
	}
}
