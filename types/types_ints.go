package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
)

// ToInt64 parse value to int64
func ToInt64(value any) (int64, error) {
	if value == nil {
		return 0, nil
	}
	switch v := value.(type) {
	case json.Number:
		return jsonNumberToInt64(v)
	case string:
		return strconv.ParseInt(v, 10, 64)
	}
	n, err := kindToInt64(value)
	if err != nil {
		if errors.Is(err, errUnsupportedKind) {
			return 0, fmt.Errorf("types: cannot convert %T to int64", value)
		}
		return 0, err
	}
	return n, nil
}

// ToInt parse value to int
func ToInt(value any) (int, error) {
	if value == nil {
		return 0, nil
	}
	switch v := value.(type) {
	case json.Number:
		i, err := jsonNumberToInt64(v)
		if err != nil {
			return 0, err
		}
		return int64ToInt(i, "json.Number")
	case string:
		return strconv.Atoi(v)
	}
	n, err := kindToInt64(value)
	if err != nil {
		if errors.Is(err, errUnsupportedKind) {
			return 0, fmt.Errorf("types: cannot convert %T to int", value)
		}
		return 0, err
	}
	return int64ToInt(n, fmt.Sprintf("%T", value))
}

func kindToInt64(value any) (int64, error) {
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uint64ToInt64(rv.Uint(), rv.Type().String())
	case reflect.Float32, reflect.Float64:
		return float64ToInt64(rv.Float(), rv.Type().String())
	default:
		return 0, errUnsupportedKind
	}
}

var errUnsupportedKind = fmt.Errorf("types: unsupported kind")

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

func jsonNumberToInt64(v json.Number) (int64, error) {
	if i, err := v.Int64(); err == nil {
		return i, nil
	}

	r, ok := new(big.Rat).SetString(v.String())
	if !ok || r.Denom().Cmp(big.NewInt(1)) != 0 || !r.Num().IsInt64() {
		return 0, fmt.Errorf("types: cannot convert json.Number %q to int64", v.String())
	}
	return r.Num().Int64(), nil
}

func float64ToInt64(v float64, typeName string) (int64, error) {
	if math.IsNaN(v) || math.IsInf(v, 0) || math.Trunc(v) != v || v < float64(math.MinInt64) || v >= -float64(math.MinInt64) {
		return 0, fmt.Errorf("types: cannot convert %s %v to int64", typeName, v)
	}
	return int64(v), nil
}
