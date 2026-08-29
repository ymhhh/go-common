package types

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"reflect"
	"strconv"
)

var _ flag.Value = (*Fund)(nil)
var _ flag.Getter = (*Fund)(nil)
var _ flag.Value = (*Found)(nil)
var _ flag.Getter = (*Found)(nil)

// Fund is a two-decimal monetary/amount value for flags and YAML.
type Fund float64

// Found is a historical alias of Fund.
type Found = Fund

// String implements flag.Value
func (p Fund) String() string {
	return fmt.Sprintf("%0.2f", p)
}

// Set implements flag.Value
func (p *Fund) Set(f string) error {
	d, err := strconv.ParseFloat(f, 64)
	if err != nil {
		return err
	}
	*p = Fund(d)
	return nil
}

// Get implements flag.Getter.
func (p Fund) Get() any {
	return float64(p)
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (p *Fund) UnmarshalYAML(unmarshal func(any) error) error {
	var f float64
	if err := unmarshal(&f); err != nil {
		return err
	}
	return p.Set(strconv.FormatFloat(f, 'f', 2, 64))
}

// MarshalYAML implements yaml.Marshaler.
func (p *Fund) MarshalYAML() (any, error) {
	if p == nil {
		return 0, nil
	}
	return *p, nil
}

// ToFloat64 converts any supported scalar type to float64.
func ToFloat64(value any) (float64, error) {
	if value == nil {
		return 0, nil
	}

	switch t := value.(type) {
	case json.Number:
		return t.Float64()
	case string:
		return strconv.ParseFloat(t, 64)
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Float32, reflect.Float64:
		return rv.Float(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(rv.Uint()), nil
	default:
		return 0, fmt.Errorf("types: cannot convert %T to float64", value)
	}
}

// RoundFund rounds fund to the nearest integer, rounding half away from zero.
func RoundFund(fund float64) int64 {
	return int64(math.Round(fund))
}
