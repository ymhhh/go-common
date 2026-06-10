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
		return uint64ToInt64(uint64(v))
	case uint64:
		return uint64ToInt64(v)
	case uint32:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case float32:
		return int64(v), nil
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
		return uint64ToInt(uint64(v))
	case uint64:
		return uint64ToInt(v)
	case uint32:
		return uint64ToInt(uint64(v))
	case uint16:
		return int(v), nil
	case uint8:
		return int(v), nil
	case float64:
		return int(v), nil
	case float32:
		return int(v), nil
	case json.Number:
		i, err := v.Int64()
		return int(i), err
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("type is valid: %s", reflect.TypeOf(value).String())
	}
}

func uint64ToInt64(v uint64) (int64, error) {
	if v > uint64(math.MaxInt64) {
		return 0, fmt.Errorf("types: %d overflows int64", v)
	}
	return int64(v), nil
}

func uint64ToInt(v uint64) (int, error) {
	if v > uint64(math.MaxInt) {
		return 0, fmt.Errorf("types: %d overflows int", v)
	}
	return int(v), nil
}
