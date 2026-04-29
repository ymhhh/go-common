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
	var (
		ds   = int64(d)
		unit = "ms"
	)
	if ds == 0 {
		return "0s"
	}

	hour := int64(time.Hour)
	factors := map[string]int64{
		"y":  hour * 24 * 365,
		"w":  hour * 24 * 7,
		"d":  hour * 24,
		"h":  hour,
		"m":  int64(time.Minute),
		"s":  int64(time.Second),
		"ms": int64(time.Millisecond),
		"us": int64(time.Microsecond),
		"ns": int64(time.Nanosecond),
	}

	switch int64(0) {
	case ds % factors["y"]:
		unit = "y"
	case ds % factors["w"]:
		unit = "w"
	case ds % factors["d"]:
		unit = "d"
	case ds % factors["h"]:
		unit = "h"
	case ds % factors["m"]:
		unit = "m"
	case ds % factors["s"]:
		unit = "s"
	case ds % factors["ms"]:
		unit = "ms"
	case ds % factors["us"]:
		unit = "us"
	case ds % factors["ns"]:
		unit = "ns"
	}
	return fmt.Sprintf("%v%v", ds/factors[unit], unit)
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
		dur := ParseStringTime(s)
		if dur == 0 && s != "" && s != "0" {
			if n, err := strconv.ParseInt(s, 10, 64); err == nil {
				*d = Duration(time.Duration(n))
				return nil
			}
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				*d = Duration(time.Duration(f))
				return nil
			}
		}
		*d = Duration(dur)
		return nil
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

	*d = 0
	return nil
}
