package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/ymhhh/go-common/types"
)

// Value wraps an arbitrary config value and provides conversions.
type Value struct {
	v any
}

func (v Value) Any() any { return v.v }

func (v Value) String() (string, error) {
	switch x := v.v.(type) {
	case nil:
		return "", fmt.Errorf("config: cannot convert <nil> to string")
	case string:
		return x, nil
	case []byte:
		return string(x), nil
	case fmt.Stringer:
		return x.String(), nil
	case bool:
		return strconv.FormatBool(x), nil
	}
	if s, ok := formatScalarNumber(v.v); ok {
		return s, nil
	}
	b, err := json.Marshal(v.v)
	if err != nil {
		return "", fmt.Errorf("config: cannot convert %T to string: %w", v.v, err)
	}
	return string(b), nil
}

func formatScalarNumber(v any) (string, bool) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Float32:
		return strconv.FormatFloat(rv.Float(), 'f', -1, 32), true
	case reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'f', -1, 64), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(rv.Uint(), 10), true
	default:
		return "", false
	}
}

func (v Value) Int() (int, error) {
	if v.v == nil {
		return 0, fmt.Errorf("config: cannot convert <nil> to int")
	}
	return types.ToInt(v.v)
}

func (v Value) Float64() (float64, error) {
	if v.v == nil {
		return 0, fmt.Errorf("config: cannot convert <nil> to float64")
	}
	f, err := types.ToFloat64(v.v)
	if err != nil {
		return 0, fmt.Errorf("config: cannot convert %T to float64: %w", v.v, err)
	}
	return f, nil
}

func (v Value) Bool() (bool, error) {
	switch x := v.v.(type) {
	case nil:
		return false, fmt.Errorf("config: cannot convert <nil> to bool")
	case bool:
		return x, nil
	case json.Number:
		i, err := x.Int64()
		return i != 0, err
	case string:
		return strconv.ParseBool(strings.TrimSpace(x))
	}
	if b, ok := boolFromNumericKind(v.v); ok {
		return b, nil
	}
	return false, fmt.Errorf("config: cannot convert %T to bool", v.v)
}

func boolFromNumericKind(v any) (bool, bool) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() != 0, true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() != 0, true
	case reflect.Float32, reflect.Float64:
		return rv.Float() != 0, true
	default:
		return false, false
	}
}

func (v Value) Map() (map[string]any, error) {
	switch x := v.v.(type) {
	case map[string]any:
		return x, nil
	case nil:
		return nil, fmt.Errorf("config: cannot convert <nil> to map")
	default:
		return nil, fmt.Errorf("config: cannot convert %T to map[string]any", v.v)
	}
}

// Slice converts the wrapped value to a []any slice.
//
// Supported inputs include JSON/YAML decoded sequences ([]any), other Go slice types
// (via reflection), and JSON array strings.
func (v Value) Slice() ([]any, error) {
	switch x := v.v.(type) {
	case nil:
		return nil, fmt.Errorf("config: cannot convert <nil> to slice")
	case []any:
		return x, nil
	case string:
		var s []any
		if err := json.Unmarshal([]byte(x), &s); err != nil {
			return nil, fmt.Errorf("config: cannot convert string to []any: %w", err)
		}
		return s, nil
	default:
		rv := reflect.ValueOf(v.v)
		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return nil, fmt.Errorf("config: cannot convert %T to []any", v.v)
		}
		n := rv.Len()
		out := make([]any, n)
		for i := range n {
			out[i] = rv.Index(i).Interface()
		}
		return out, nil
	}
}
