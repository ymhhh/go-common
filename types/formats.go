package types

import (
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	timeReg = `^(?P<value>([0-9]+(\.[0-9]+)?))\s*(?P<unit>(nanoseconds|nanosecond|nanos|nano|ns|microseconds|microsecond|micros|micro|us|milliseconds|millisecond|millis|milli|ms|seconds|second|s|minutes|minute|m|hours|hour|h|days|day|d|weeks|week|w|years|year|y))$`

	bitReg = `^(?P<value>([0-9]+(\.[0-9]+)?))\s*(?P<unit>(b|byte|bytes|kb|kilobyte|kilobytes|mb|megabyte|megabytes|gb|gigabyte|gigabytes|tb|terabyte|terabytes|pb|petabyte|petabytes|eb|exabyte|exabytes|zb|zettabyte|zettabytes|yb|yottabyte|yottabytes|k|ki|kib|kibibyte|kibibytes|m|mi|mib|mebibyte|mebibytes|g|gi|gib|gibibyte|gibibytes|t|ti|tib|tebibyte|tebibytes|p|pi|pib|pebibyte|pebibytes|e|ei|eib|exbibyte|exbibytes|z|zi|zib|zebibyte|zebibytes|y|yi|yib|yobibyte|yobibytes))$`
)

var (
	timeRe = regexp.MustCompile(timeReg)
	bitRe  = regexp.MustCompile(bitReg)
)

// ByteSizes
var (
	_Num1000 = big.NewInt(1000)
	_Num1024 = big.NewInt(1024)

	_Byte   = big.NewInt(1)
	_KiByte = (&big.Int{}).Mul(_Byte, _Num1024)
	_MiByte = (&big.Int{}).Mul(_KiByte, _Num1024)
	_GiByte = (&big.Int{}).Mul(_MiByte, _Num1024)
	_TiByte = (&big.Int{}).Mul(_GiByte, _Num1024)
	_PiByte = (&big.Int{}).Mul(_TiByte, _Num1024)
	_EiByte = (&big.Int{}).Mul(_PiByte, _Num1024)
	_ZiByte = (&big.Int{}).Mul(_EiByte, _Num1024)
	_YiByte = (&big.Int{}).Mul(_ZiByte, _Num1024)

	_KByte = (&big.Int{}).Mul(_Byte, _Num1000)
	_MByte = (&big.Int{}).Mul(_KByte, _Num1000)
	_GByte = (&big.Int{}).Mul(_MByte, _Num1000)
	_TByte = (&big.Int{}).Mul(_GByte, _Num1000)
	_PByte = (&big.Int{}).Mul(_TByte, _Num1000)
	_EByte = (&big.Int{}).Mul(_PByte, _Num1000)
	_ZByte = (&big.Int{}).Mul(_EByte, _Num1000)
	_YByte = (&big.Int{}).Mul(_ZByte, _Num1000)
)

var byteSizeUnits = func() map[string]*big.Int {
	m := make(map[string]*big.Int, 64)
	add := func(unit *big.Int, names ...string) {
		for _, name := range names {
			m[name] = unit
		}
	}
	add(_Byte, "b", "byte", "bytes")
	add(_KByte, "kb", "kilobyte", "kilobytes")
	add(_MByte, "mb", "megabyte", "megabytes")
	add(_GByte, "gb", "gigabyte", "gigabytes")
	add(_TByte, "tb", "terabyte", "terabytes")
	add(_PByte, "pb", "petabyte", "petabytes")
	add(_EByte, "eb", "exabyte", "exabytes")
	add(_ZByte, "zb", "zettabyte", "zettabytes")
	add(_YByte, "yb", "yottabyte", "yottabytes")
	add(_KiByte, "k", "ki", "kib", "kibibyte", "kibibytes")
	add(_MiByte, "m", "mi", "mib", "mebibyte", "mebibytes")
	add(_GiByte, "g", "gi", "gib", "gibibyte", "gibibytes")
	add(_TiByte, "t", "ti", "tib", "tebibyte", "tebibytes")
	add(_PiByte, "p", "pi", "pib", "pebibyte", "pebibytes")
	add(_EiByte, "e", "ei", "eib", "exbibyte", "exbibytes")
	add(_ZiByte, "z", "zi", "zib", "zebibyte", "zebibytes")
	add(_YiByte, "y", "yi", "yib", "yobibyte", "yobibytes")
	return m
}()

var timeUnits = map[string]time.Duration{
	"nanoseconds":  time.Nanosecond,
	"nanosecond":   time.Nanosecond,
	"nanos":        time.Nanosecond,
	"nano":         time.Nanosecond,
	"ns":           time.Nanosecond,
	"microseconds": time.Microsecond,
	"microsecond":  time.Microsecond,
	"micros":       time.Microsecond,
	"micro":        time.Microsecond,
	"us":           time.Microsecond,
	"milliseconds": time.Millisecond,
	"millisecond":  time.Millisecond,
	"millis":       time.Millisecond,
	"milli":        time.Millisecond,
	"ms":           time.Millisecond,
	"seconds":      time.Second,
	"second":       time.Second,
	"s":            time.Second,
	"minutes":      time.Minute,
	"minute":       time.Minute,
	"m":            time.Minute,
	"hours":        time.Hour,
	"hour":         time.Hour,
	"h":            time.Hour,
	"days":         24 * time.Hour,
	"day":          24 * time.Hour,
	"d":            24 * time.Hour,
	"weeks":        7 * 24 * time.Hour,
	"week":         7 * 24 * time.Hour,
	"w":            7 * 24 * time.Hour,
	"years":        365 * 24 * time.Hour,
	"year":         365 * 24 * time.Hour,
	"y":            365 * 24 * time.Hour,
}

// FindStringSubmatchMap returns a map of named capture groups from the leftmost match
// of re in s. A return value of nil indicates no match.
func FindStringSubmatchMap(s string, re *regexp.Regexp) (map[string]string, bool) {
	captures := make(map[string]string)

	match := re.FindStringSubmatch(s)
	if match == nil {
		return captures, false
	}

	for i, name := range re.SubexpNames() {
		if i == 0 || name == "" {
			continue
		}
		captures[name] = match[i]
	}
	return captures, true
}

// ParseStringByteSize return big size
func ParseStringByteSize(key string, defValue ...*big.Int) *big.Int {
	groups, matched := FindStringSubmatchMap(strings.ToLower(key), bitRe)
	if !matched {
		return defaultByteSize(defValue...)
	}
	unit, ok := byteSizeUnits[groups["unit"]]
	if !ok {
		return defaultByteSize(defValue...)
	}
	return parseByteSizeValue(groups["value"], unit, defValue...)
}

func defaultByteSize(defValue ...*big.Int) *big.Int {
	if len(defValue) == 0 {
		return nil
	}
	return defValue[0]
}

func parseByteSizeValue(value string, unit *big.Int, defValue ...*big.Int) *big.Int {
	r, ok := new(big.Rat).SetString(value)
	if !ok {
		return defaultByteSize(defValue...)
	}
	r.Mul(r, new(big.Rat).SetInt(unit))
	// Integer division truncates toward zero (e.g. 1.3kib → 1331, not 1331.2).
	return new(big.Int).Quo(r.Num(), r.Denom())
}

// ParseStringTimeOK parses a human-readable duration with a unit
// (e.g. "2s", "0d", "1.5h"). ok is false when s does not match.
// A successful parse of zero (e.g. "0d", "0s") returns (0, true).
func ParseStringTimeOK(s string) (time.Duration, bool) {
	groups, matched := FindStringSubmatchMap(strings.ToLower(s), timeRe)
	if !matched {
		return 0, false
	}

	f, err := strconv.ParseFloat(groups["value"], 64)
	if err != nil {
		return 0, false
	}
	unit, ok := timeUnits[groups["unit"]]
	if !ok {
		return 0, false
	}
	return time.Duration(float64(unit) * f), true
}

// ParseStringTime return time.Duration.
// When s does not match, returns defValue[0] if provided, otherwise 0.
// Prefer ParseStringTimeOK when zero must be distinguished from parse failure.
func ParseStringTime(s string, defValue ...time.Duration) time.Duration {
	d, ok := ParseStringTimeOK(s)
	if ok {
		return d
	}
	if len(defValue) == 0 {
		return 0
	}
	return defValue[0]
}
