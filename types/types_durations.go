package types

import (
	"flag"
	"fmt"
	"strconv"
	"time"
)

// Duration wraps time.Duration and implements flag.Value.
//
// Example:
//
//	var d types.Duration
//	flag.Var(&d, "timeout", "request timeout (e.g. 200ms, 3s, 1m)")
type Duration time.Duration

var _ flag.Value = (*Duration)(nil)
var _ flag.Getter = (*Duration)(nil)

func (d Duration) String() string {
	ds := int64(d)
	if ds == 0 {
		return "0s"
	}

	units := []struct {
		name   string
		factor int64
	}{
		{"y", int64(time.Hour * 24 * 365)},
		{"w", int64(time.Hour * 24 * 7)},
		{"d", int64(time.Hour * 24)},
		{"h", int64(time.Hour)},
		{"m", int64(time.Minute)},
		{"s", int64(time.Second)},
		{"ms", int64(time.Millisecond)},
		{"us", int64(time.Microsecond)},
		{"ns", int64(time.Nanosecond)},
	}

	for _, u := range units {
		if ds%u.factor == 0 {
			return fmt.Sprintf("%v%v", ds/u.factor, u.name)
		}
	}
	return fmt.Sprintf("%vns", ds)
}

func (d *Duration) Set(s string) error {
	if s == "" {
		*d = 0
		return nil
	}
	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(v)
	return nil
}

func (d Duration) Get() any {
	return time.Duration(d)
}

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

func (d Duration) MarshalYAML() (any, error) {
	return d.String(), nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (d *Duration) UnmarshalYAML(unmarshal func(any) error) error {
	var s string
	if err := unmarshal(&s); err == nil {
		if s == "" {
			*d = 0
			return nil
		}
		dur := ParseStringTime(s)
		if dur != 0 || s == "0" {
			*d = Duration(dur)
			return nil
		}
		if d2, err := time.ParseDuration(s); err == nil {
			*d = Duration(d2)
			return nil
		}
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			*d = Duration(time.Duration(n))
			return nil
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			*d = Duration(time.Duration(f))
			return nil
		}
		return fmt.Errorf("types: invalid duration %q", s)
	}

	var n int64
	if err := unmarshal(&n); err == nil {
		*d = Duration(time.Duration(n))
		return nil
	}

	var f float64
	if err := unmarshal(&f); err == nil {
		*d = Duration(time.Duration(f))
		return nil
	}

	return fmt.Errorf("types: invalid duration value")
}
