package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
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
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64), nil
	case float32:
		return strconv.FormatFloat(float64(x), 'f', -1, 32), nil
	case int:
		return strconv.Itoa(x), nil
	case int64:
		return strconv.FormatInt(x, 10), nil
	case int32:
		return strconv.FormatInt(int64(x), 10), nil
	case uint:
		return strconv.FormatUint(uint64(x), 10), nil
	case uint64:
		return strconv.FormatUint(x, 10), nil
	case bool:
		if x {
			return "true", nil
		}
		return "false", nil
	default:
		b, err := json.Marshal(x)
		if err != nil {
			return "", fmt.Errorf("config: cannot convert %T to string: %w", v.v, err)
		}
		return string(b), nil
	}
}

func (v Value) Int() (int, error) {
	switch x := v.v.(type) {
	case int:
		return x, nil
	case int64:
		return int(x), nil
	case int32:
		return int(x), nil
	case uint:
		return int(x), nil
	case uint64:
		return int(x), nil
	case float64:
		return int(x), nil
	case float32:
		return int(x), nil
	case json.Number:
		i, err := x.Int64()
		return int(i), err
	case string:
		i, err := strconv.ParseInt(x, 10, 64)
		return int(i), err
	default:
		return 0, fmt.Errorf("config: cannot convert %T to int", v.v)
	}
}

func (v Value) Float64() (float64, error) {
	switch x := v.v.(type) {
	case float64:
		return x, nil
	case float32:
		return float64(x), nil
	case int:
		return float64(x), nil
	case int64:
		return float64(x), nil
	case int32:
		return float64(x), nil
	case uint:
		return float64(x), nil
	case uint64:
		return float64(x), nil
	case json.Number:
		return x.Float64()
	case string:
		return strconv.ParseFloat(x, 64)
	default:
		return 0, fmt.Errorf("config: cannot convert %T to float64", v.v)
	}
}

func (v Value) Bool() (bool, error) {
	switch x := v.v.(type) {
	case nil:
		return false, fmt.Errorf("config: cannot convert <nil> to bool")
	case bool:
		return x, nil
	case int:
		return x != 0, nil
	case int64:
		return x != 0, nil
	case int32:
		return x != 0, nil
	case uint:
		return x != 0, nil
	case uint64:
		return x != 0, nil
	case float64:
		return x != 0, nil
	case float32:
		return x != 0, nil
	case json.Number:
		i, err := x.Int64()
		return i != 0, err
	case string:
		return strconv.ParseBool(strings.TrimSpace(x))
	default:
		return false, fmt.Errorf("config: cannot convert %T to bool", v.v)
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
