package types

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
)

// ToInt64 parse value to int64
func ToInt64(value any) (int64, error) {
	if value == nil {
		return 0, nil
	}
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case uint:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case float64:
		return float64ToInt64(v, "float64")
	case float32:
		return float64ToInt64(float64(v), "float32")
	case json.Number:
		return v.Int64()
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("type is valid: %s", reflect.TypeOf(value).String())
	}
}

// ToInt parse value to int
func ToInt(value any) (int, error) {
	if value == nil {
		return 0, nil
	}
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case int32:
		return int(v), nil
	case int16:
		return int(v), nil
	case int8:
		return int(v), nil
	case uint:
		return int(v), nil
	case uint64:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint8:
		return int(v), nil
	case float64:
		return float64ToInt(v, "float64")
	case float32:
		return float64ToInt(float64(v), "float32")
	case json.Number:
		i, err := v.Int64()
		return int(i), err
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("type is valid: %s", reflect.TypeOf(value).String())
	}
}

func float64ToInt64(v float64, typeName string) (int64, error) {
	if math.IsNaN(v) || math.IsInf(v, 0) || math.Trunc(v) != v || v < float64(math.MinInt64) || v >= -float64(math.MinInt64) {
		return 0, fmt.Errorf("types: cannot convert %s %v to int64", typeName, v)
	}
	return int64(v), nil
}

func float64ToInt(v float64, typeName string) (int, error) {
	if math.IsNaN(v) || math.IsInf(v, 0) || math.Trunc(v) != v || v < float64(math.MinInt) || v >= -float64(math.MinInt) {
		return 0, fmt.Errorf("types: cannot convert %s %v to int", typeName, v)
	}
	return int(v), nil
}
