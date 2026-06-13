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
		return uint64ToInt64(uint64(v), "uint")
	case uint64:
		return uint64ToInt64(v, "uint64")
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
		return int64ToInt(v, "int64")
	case int32:
		return int64ToInt(int64(v), "int32")
	case int16:
		return int(v), nil
	case int8:
		return int(v), nil
	case uint:
		return uint64ToInt(uint64(v), "uint")
	case uint64:
		return uint64ToInt(v, "uint64")
	case uint32:
		return uint64ToInt(uint64(v), "uint32")
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
		if err != nil {
			return 0, err
		}
		return int64ToInt(i, "json.Number")
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("type is valid: %s", reflect.TypeOf(value).String())
	}
}

func uint64ToInt64(v uint64, typeName string) (int64, error) {
	if v > uint64(math.MaxInt64) {
		return 0, fmt.Errorf("types: cannot convert %s %d to int64", typeName, v)
	}
	return int64(v), nil
}

func int64ToInt(v int64, typeName string) (int, error) {
	if v < int64(math.MinInt) || v > int64(math.MaxInt) {
		return 0, fmt.Errorf("types: cannot convert %s %d to int", typeName, v)
	}
	return int(v), nil
}

func uint64ToInt(v uint64, typeName string) (int, error) {
	if v > uint64(math.MaxInt) {
		return 0, fmt.Errorf("types: cannot convert %s %d to int", typeName, v)
	}
	return int(v), nil
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
