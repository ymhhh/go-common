package types

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"reflect"
	"strconv"
)

var _ flag.Value = (*Found)(nil)
var _ flag.Getter = (*Found)(nil)

// Found represents a found value with precision up to two decimal places.
type Found float64

// String implements flag.Value
func (p Found) String() string {
	return fmt.Sprintf("%0.2f", p)
}

// Set implements flag.Value
func (p *Found) Set(f string) error {
	d, err := strconv.ParseFloat(f, 64)
	if err != nil {
		return err
	}
	*p = Found(d)
	return nil
}

// Get implements flag.Getter.
func (p Found) Get() any {
	return float64(p)
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (p *Found) UnmarshalYAML(unmarshal func(any) error) error {
	var f float64
	if err := unmarshal(&f); err != nil {
		return err
	}
	return p.Set(strconv.FormatFloat(f, 'f', 2, 64))
}

// MarshalYAML implements yaml.Marshaler.
func (p *Found) MarshalYAML() (any, error) {
	if p == nil {
		return 0, nil
	}
	return *p, nil
}

// ToFloat64 covert any type to float64
func ToFloat64(value any) (float64, error) {
	if value == nil {
		return 0, nil
	}

	switch t := value.(type) {
	case float64:
		return t, nil
	case float32:
		return float64(t), nil
	case int:
		return float64(t), nil
	case int32:
		return float64(t), nil
	case int64:
		return float64(t), nil
	case int8:
		return float64(t), nil
	case int16:
		return float64(t), nil
	case string:
		return strconv.ParseFloat(t, 64)
	case json.Number:
		return t.Float64()
	case Found:
		return float64(t), nil
	default:
		return 0, fmt.Errorf("type is valid: %s", reflect.TypeOf(value).String())
	}
}

// RoundFund round fund to int64
func RoundFund(fund float64) int64 {
	fInt, fFloat := math.Modf(fund)
	f := int64(fInt)
	if fFloat >= 0.50000000000 {
		f++
	}
	return f
}
