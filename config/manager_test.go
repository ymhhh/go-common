package config

import (
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
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

func TestManager_GetIntRejectsUnsafeNumericConversions(t *testing.T) {
	tooLarge := uint64(^uint(0)>>1) + 1
	c := newTestTree(map[string]any{
		"fractional": float64(1.5),
		"overflow":   tooLarge,
		"list":       []any{float64(2), float64(2.5), tooLarge},
	})

	if got := c.GetInt("fractional", 99); got != 99 {
		t.Fatalf("fractional value should fall back to default, got %d", got)
	}
	if got := c.GetInt("overflow", 99); got != 99 {
		t.Fatalf("overflowing value should fall back to default, got %d", got)
	}
	if got := c.GetIntList("list"); !reflect.DeepEqual(got, []int{2}) {
		t.Fatalf("unsafe list values should be skipped, got %#v", got)
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
		"d":          "500ms",
		"s":          "2kb",
		"n":          int64(1_000_000),
		"json_exp":   json.Number("1e6"),
		"fractional": float64(0.5),
	})

	if c.GetTimeDuration("d") != 500*time.Millisecond {
		t.Fatalf("duration: %v", c.GetTimeDuration("d"))
	}
	if c.GetTimeDuration("n") != time.Millisecond {
		t.Fatalf("duration ns int: %v", c.GetTimeDuration("n"))
	}
	if got := c.GetTimeDuration("json_exp"); got != time.Millisecond {
		t.Fatalf("duration ns json exponent: %v", got)
	}
	if got := c.GetTimeDuration("fractional", time.Second); got != time.Second {
		t.Fatalf("fractional numeric duration should fall back to default, got %v", got)
	}

	bs := c.GetByteSize("s")
	want := (&big.Int{}).Mul(big.NewInt(2), big.NewInt(1000))
	if bs == nil || bs.Cmp(want) != 0 {
		t.Fatalf("bytesize: got=%v want=%v", bs, want)
	}
}

func TestManager_GetByteSize_JSONNumber(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{
		"limit": 1048576,
		"huge_limit": 9223372036854775808,
		"exponent_limit": 1e6
	}`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	c, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got, want := c.GetByteSize("limit"), big.NewInt(1048576); got == nil || got.Cmp(want) != 0 {
		t.Fatalf("limit: got=%v want=%v", got, want)
	}

	hugeWant, ok := new(big.Int).SetString("9223372036854775808", 10)
	if !ok {
		t.Fatalf("invalid huge test value")
	}
	if got := c.GetByteSize("huge_limit"); got == nil || got.Cmp(hugeWant) != 0 {
		t.Fatalf("huge_limit: got=%v want=%v", got, hugeWant)
	}

	if got, want := c.GetByteSize("exponent_limit"), big.NewInt(1_000_000); got == nil || got.Cmp(want) != 0 {
		t.Fatalf("exponent_limit: got=%v want=%v", got, want)
	}
}

func TestManager_GetByteSizeRejectsUnsafeNumericConversions(t *testing.T) {
	hugeWant, ok := new(big.Int).SetString("100000000000000000000", 10)
	if !ok {
		t.Fatalf("invalid huge test value")
	}
	exactHugeWant, ok := new(big.Int).SetString("1"+strings.Repeat("0", 200), 10)
	if !ok {
		t.Fatalf("invalid exact huge test value")
	}
	def := big.NewInt(99)
	c := newTestTree(map[string]any{
		"fractional":              float64(1.5),
		"json_fractional":         json.Number("1.5"),
		"json_fractional_precise": json.Number("1." + strings.Repeat("0", 100) + "1"),
		"huge":                    float64(1e20),
		"json_huge":               json.Number("1e200"),
	})

	if got := c.GetByteSize("fractional", def); got == nil || got.Cmp(def) != 0 {
		t.Fatalf("fractional byte size should fall back to default, got %v", got)
	}
	if got := c.GetByteSize("json_fractional", def); got == nil || got.Cmp(def) != 0 {
		t.Fatalf("fractional JSON byte size should fall back to default, got %v", got)
	}
	if got := c.GetByteSize("json_fractional_precise", def); got == nil || got.Cmp(def) != 0 {
		t.Fatalf("precise fractional JSON byte size should fall back to default, got %v", got)
	}
	if got := c.GetByteSize("huge"); got == nil || got.Cmp(hugeWant) != 0 {
		t.Fatalf("huge byte size should not narrow through int64: got=%v want=%v", got, hugeWant)
	}
	if got := c.GetByteSize("json_huge"); got == nil || got.Cmp(exactHugeWant) != 0 {
		t.Fatalf("huge JSON byte size should preserve exact decimal value: got=%v want=%v", got, exactHugeWant)
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

func TestManager_MutableGettersReturnCopies(t *testing.T) {
	tr := &Tree{
		root: map[string]any{
			"items": []any{
				map[string]any{"enabled": true},
				"keep",
			},
			"section": map[string]any{
				"limit":  "10",
				"nested": map[string]any{"v": 1},
			},
		},
	}
	var c Config = tr

	list := c.GetList("items")
	list[0].(map[string]any)["enabled"] = false
	list[1] = "changed"
	if got := c.GetList("items"); got[1] != "keep" || got[0].(map[string]any)["enabled"] != true {
		t.Fatalf("GetList returned mutable config internals: %#v", got)
	}

	iface := c.GetInterface("section").(map[string]any)
	iface["limit"] = "99"
	iface["nested"].(map[string]any)["v"] = 2
	if c.GetString("section.limit") != "10" || c.GetInt("section.nested.v") != 1 {
		t.Fatalf("GetInterface returned mutable config internals: limit=%q nested=%d", c.GetString("section.limit"), c.GetInt("section.nested.v"))
	}

	fromGet, err := c.Get("section").Map()
	if err != nil {
		t.Fatalf("Get(section).Map: %v", err)
	}
	fromGet["limit"] = "100"
	if c.GetString("section.limit") != "10" {
		t.Fatalf("Get(section).Map returned mutable config internals: %q", c.GetString("section.limit"))
	}

	sub := c.GetValuesConfig("section")
	if err := sub.Set("nested.v", 3); err != nil {
		t.Fatalf("mutate GetValuesConfig result: %v", err)
	}
	if c.GetInt("section.nested.v") != 1 {
		t.Fatalf("GetValuesConfig returned mutable config internals: %d", c.GetInt("section.nested.v"))
	}
}

func TestSet_DoesNotAliasMutableInputs(t *testing.T) {
	tr := &Tree{root: map[string]any{}}
	patch := map[string]any{
		"enabled": true,
		"rules":   []any{"deny"},
		"nested":  map[string]any{"limit": 10},
	}

	if err := tr.Set("auth", patch); err != nil {
		t.Fatalf("Set: %v", err)
	}

	patch["enabled"] = false
	patch["rules"].([]any)[0] = "allow"
	patch["nested"].(map[string]any)["limit"] = 99

	if got := tr.GetBoolean("auth.enabled"); !got {
		t.Fatalf("Set stored caller-owned map by reference: enabled=%v", got)
	}
	if got := tr.GetList("auth.rules"); len(got) != 1 || got[0] != "deny" {
		t.Fatalf("Set stored caller-owned slice by reference: %#v", got)
	}
	if got := tr.GetInt("auth.nested.limit"); got != 10 {
		t.Fatalf("Set stored caller-owned nested map by reference: %d", got)
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
