package config

import (
	"encoding/json"
	"math/big"
	"reflect"
	"testing"
	"time"
)

func newTestTree(m map[string]any) Config {
	return &Tree{root: m}
}

func TestManager_GetPrimitives_Defaults(t *testing.T) {
	c := newTestTree(map[string]any{})

	if got := c.GetInterface("missing", "def"); got != "def" {
		t.Fatalf("GetInterface: got %#v", got)
	}
	if got := c.GetString("missing", "x"); got != "x" {
		t.Fatalf("GetString default: %q", got)
	}
	if got := c.GetBoolean("missing", true); !got {
		t.Fatalf("GetBoolean default")
	}
	if got := c.GetInt("missing", 9); got != 9 {
		t.Fatalf("GetInt default: %d", got)
	}
	if got := c.GetFloat("missing", 1.25); got != 1.25 {
		t.Fatalf("GetFloat default: %v", got)
	}
	if got := c.GetTimeDuration("missing", time.Second); got != time.Second {
		t.Fatalf("GetTimeDuration default: %v", got)
	}
	def := big.NewInt(123)
	if got := c.GetByteSize("missing", def); got.Cmp(def) != 0 {
		t.Fatalf("GetByteSize default: %v", got)
	}
}

func TestManager_GetPrimitives_Conversions(t *testing.T) {
	c := newTestTree(map[string]any{
		"b": true,
		"i": float64(7),
		"f": "2.5",
	})

	if !c.GetBoolean("b") {
		t.Fatalf("bool")
	}
	if c.GetInt("i") != 7 {
		t.Fatalf("int from float: %d", c.GetInt("i"))
	}
	if c.GetFloat("f") != 2.5 {
		t.Fatalf("float from string: %v", c.GetFloat("f"))
	}
}

func TestManager_Lists(t *testing.T) {
	c := newTestTree(map[string]any{
		// avoid 0/1 integers: Value.Bool treats non-zero numbers as true, which is surprising for list typing tests
		"xs": []any{2, "3", 0, true},
	})

	if got := c.GetList("xs"); len(got) != 4 {
		t.Fatalf("list len: %d", len(got))
	}
	if got := c.GetIntList("xs"); !reflect.DeepEqual(got, []int{2, 3, 0}) {
		t.Fatalf("int list: %#v", got)
	}
	if got := c.GetFloatList("xs"); len(got) != 3 || got[0] != 2 || got[1] != 3 || got[2] != 0 {
		t.Fatalf("float list: %#v", got)
	}
	if got := c.GetBooleanList("xs"); !reflect.DeepEqual(got, []bool{true, false, true}) {
		t.Fatalf("bool list: %#v", got)
	}
	if got := c.GetStringList("xs"); len(got) != 4 {
		t.Fatalf("string list len: %d", len(got))
	}
}

func TestManager_TimeAndByteSize(t *testing.T) {
	c := newTestTree(map[string]any{
		"d": "500ms",
		"s": "2kb",
		"n": int64(1_000_000),
	})

	if c.GetTimeDuration("d") != 500*time.Millisecond {
		t.Fatalf("duration: %v", c.GetTimeDuration("d"))
	}
	if c.GetTimeDuration("n") != time.Millisecond {
		t.Fatalf("duration ns int: %v", c.GetTimeDuration("n"))
	}

	bs := c.GetByteSize("s")
	want := (&big.Int{}).Mul(big.NewInt(2), big.NewInt(1000))
	if bs == nil || bs.Cmp(want) != 0 {
		t.Fatalf("bytesize: got=%v want=%v", bs, want)
	}
}

func TestManager_Map_Subconfig_Copy_Dump(t *testing.T) {
	tr := &Tree{
		root: map[string]any{
			"a": map[string]any{
				"k": float64(1),
			},
		},
	}
	var c Config = tr

	m := c.GetMap("a")
	if m == nil || m["k"] != float64(1) {
		t.Fatalf("GetMap: %#v", m)
	}

	sub := c.GetConfig("a")
	if sub.GetInt("k") != 1 {
		t.Fatalf("GetConfig: %d", sub.GetInt("k"))
	}

	cp := c.Copy()
	if err := cp.Set("a.k", 2); err != nil {
		t.Fatalf("mutate copy: %v", err)
	}
	if c.GetInt("a.k") != 1 {
		t.Fatalf("copy should be deep: orig=%d", c.GetInt("a.k"))
	}

	bs, err := c.Dump()
	if err != nil {
		t.Fatalf("Dump: %v", err)
	}
	var round map[string]any
	if err := json.Unmarshal(bs, &round); err != nil {
		t.Fatalf("Dump json: %v", err)
	}
	if !reflect.DeepEqual(round, tr.root) {
		t.Fatalf("dump roundtrip mismatch")
	}
}

func TestManager_Object(t *testing.T) {
	type obj struct {
		N int `json:"n"`
	}
	c := newTestTree(map[string]any{
		"x": map[string]any{
			"n": 9,
		},
	})

	var out obj
	if err := c.Object(&out, WithObjectPath("x")); err != nil {
		t.Fatalf("Object: %v", err)
	}
	if out.N != 9 {
		t.Fatalf("n=%d", out.N)
	}
}

func TestManager_ToObjectCompatibility(t *testing.T) {
	type obj struct {
		N int `json:"n"`
	}
	c := newTestTree(map[string]any{
		"x": map[string]any{
			"n": 9,
		},
	})

	var out obj
	if err := c.ToObject("x", &out); err != nil {
		t.Fatalf("ToObject: %v", err)
	}
	if out.N != 9 {
		t.Fatalf("n=%d", out.N)
	}
}

func TestManager_GetValuesConfig_Panic(t *testing.T) {
	c := newTestTree(map[string]any{
		"bad": 1,
	})

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()
	_ = c.GetValuesConfig("bad")
}

func TestManager_GetRootKeys_IsEmpty(t *testing.T) {
	if !newTestTree(map[string]any{}).IsEmpty() {
		t.Fatalf("expected empty")
	}

	c := newTestTree(map[string]any{"z": 1, "a": 2})
	keys := c.GetRootKeys()
	if len(keys) != 2 || keys[0] != "a" || keys[1] != "z" {
		t.Fatalf("keys: %#v", keys)
	}
}
