package config

import (
	"reflect"
	"testing"
)

func TestDeepCopy_map(t *testing.T) {
	src := map[string]any{
		"a": 1,
		"b": map[string]any{
			"c": []any{1, 2, map[string]any{"d": 3}},
		},
	}
	cp := DeepCopy(src).(map[string]any)
	if reflect.ValueOf(cp).Pointer() == reflect.ValueOf(src).Pointer() {
		t.Fatal("top map: same pointer")
	}
	inner := cp["b"].(map[string]any)
	origInner := src["b"].(map[string]any)
	if reflect.ValueOf(inner).Pointer() == reflect.ValueOf(origInner).Pointer() {
		t.Fatal("nested map: same pointer")
	}
	sl := inner["c"].([]any)
	origSl := origInner["c"].([]any)
	nm := sl[2].(map[string]any)
	origNM := origSl[2].(map[string]any)
	if reflect.ValueOf(nm).Pointer() == reflect.ValueOf(origNM).Pointer() {
		t.Fatal("nested map in slice: same pointer")
	}

	cp["a"] = 99
	if src["a"] != 1 {
		t.Fatalf("mutate copy affected src: %v", src["a"])
	}
}

func TestDeepCopy_mapAnyAny(t *testing.T) {
	src := map[any]any{
		"a": 1,
		"b": map[any]any{
			"c": []any{1, 2, map[any]any{"d": 3}},
		},
	}
	cp := DeepCopy(src).(map[any]any)
	if reflect.ValueOf(cp).Pointer() == reflect.ValueOf(src).Pointer() {
		t.Fatal("top map: same pointer")
	}
	inner := cp["b"].(map[any]any)
	origInner := src["b"].(map[any]any)
	if reflect.ValueOf(inner).Pointer() == reflect.ValueOf(origInner).Pointer() {
		t.Fatal("nested map: same pointer")
	}
	sl := inner["c"].([]any)
	origSl := origInner["c"].([]any)
	nm := sl[2].(map[any]any)
	origNM := origSl[2].(map[any]any)
	if reflect.ValueOf(nm).Pointer() == reflect.ValueOf(origNM).Pointer() {
		t.Fatal("nested map in slice: same pointer")
	}
	cp["a"] = 99
	if src["a"] != 1 {
		t.Fatalf("mutate copy affected src: %v", src["a"])
	}
}

func TestDeepCopy_nil(t *testing.T) {
	if DeepCopy(nil) != nil {
		t.Fatal("expected nil")
	}
}

func TestDeepCopy_leaf(t *testing.T) {
	if DeepCopy(42) != 42 {
		t.Fatal("int leaf")
	}
	if DeepCopy("x") != "x" {
		t.Fatal("string leaf")
	}
}

func TestDeepCopy_empty(t *testing.T) {
	emptyStrMap := map[string]any{}
	cp1 := DeepCopy(emptyStrMap).(map[string]any)
	if reflect.ValueOf(cp1).Pointer() == reflect.ValueOf(emptyStrMap).Pointer() {
		t.Fatal("empty map[string]any: expected new map")
	}
	cp1["k"] = 1
	if len(emptyStrMap) != 0 {
		t.Fatal("mutating copy grew original empty map")
	}

	emptyAnyMap := map[any]any{}
	cp2 := DeepCopy(emptyAnyMap).(map[any]any)
	if reflect.ValueOf(cp2).Pointer() == reflect.ValueOf(emptyAnyMap).Pointer() {
		t.Fatal("empty map[any]any: expected new map")
	}

	emptySl := []any{}
	cp3 := DeepCopy(emptySl).([]any)
	if len(cp3) != 0 {
		t.Fatalf("empty []any copy: len %d", len(cp3))
	}
	cp3 = append(cp3, 1)
	if len(emptySl) != 0 {
		t.Fatal("append to copy of empty slice affected original")
	}
}

func TestDeepCopy_stringMap_nestedMapAnyAny(t *testing.T) {
	// JSON/YAML 解码后常见：外层 map[string]any，内层对象偶尔为 map[any]any
	src := map[string]any{
		"outer": map[any]any{
			"n": 42,
			"m": map[string]any{"x": 1},
		},
	}
	cp := DeepCopy(src).(map[string]any)
	inner := cp["outer"].(map[any]any)
	origInner := src["outer"].(map[any]any)
	if reflect.ValueOf(inner).Pointer() == reflect.ValueOf(origInner).Pointer() {
		t.Fatal("nested map[any]any: same pointer")
	}
	sub := inner["m"].(map[string]any)
	origSub := origInner["m"].(map[string]any)
	if reflect.ValueOf(sub).Pointer() == reflect.ValueOf(origSub).Pointer() {
		t.Fatal("nested map[string]any: same pointer")
	}
	sub["x"] = 99
	if origSub["x"] != 1 {
		t.Fatalf("copy mutation leaked: %v", origSub["x"])
	}
}

func TestDeepCopy_mapAnyAny_nonStringKeys(t *testing.T) {
	src := map[any]any{
		1:     "a",
		"two": 2,
		3.0:   true,
	}
	cp := DeepCopy(src).(map[any]any)
	if len(cp) != len(src) {
		t.Fatalf("len: got %d want %d", len(cp), len(src))
	}
	if cp[1] != "a" || cp["two"] != 2 || cp[3.0] != true {
		t.Fatalf("values: %#v", cp)
	}
	cp[1] = "b"
	if src[1] != "a" {
		t.Fatalf("mutate int key: src changed to %v", src[1])
	}
}
