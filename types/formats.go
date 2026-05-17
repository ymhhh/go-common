package types

import (
	"math/big"
	"regexp"
	"time"
)

const (
	timeReg = `^(?P<value>([0-9]+(\.[0-9]+)?))\s*(?P<unit>(nanoseconds|nanosecond|nanos|nano|ns|microseconds|microsecond|micros|micro|us|milliseconds|millisecond|millis|milli|ms|seconds|second|s|minutes|minute|m|hours|hour|h|days|day|d|weeks|week|w|years|year|y))$`

	bitReg = `^(?P<value>([0-9]+(\.[0-9]+)?))\s*(?P<unit>(b|byte|bytes|kb|kilobyte|kilobytes|mb|megabyte|megabytes|gb|gigabyte|gigabytes|tb|terabyte|terabytes|pb|petabyte|petabytes|eb|exabyte|exabytes|zb|zettabyte|zettabytes|yb|yottabyte|yottabytes|k|ki|kib|kibibyte|kibibytes|m|mi|mib|mebibyte|mebibytes|g|gi|gib|gibibyte|gibibytes|t|ti|tib|tebibyte|tebibytes|p|pi|pib|pebibyte|pebibytes|e|ei|eib|exbibyte|exbibytes|z|zi|zib|zebibyte|zebibytes|y|yi|yib|yobibyte|yobibytes))$`
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

// FindStringSubmatchMap information:
// returns a map of strings holding the text of the
// leftmost match of the regular expression in s and the matches, if any, of
// its subexpressions, as defined by the 'SubMatch' description in the
// package comment.
// A return value of nil indicates no match.
func FindStringSubmatchMap(s, exp string) (map[string]string, bool) {
	reg := regexp.MustCompile(exp)
	captures := make(map[string]string)

	match := reg.FindStringSubmatch(s)
	if match == nil {
		return captures, false
	}

	for i, name := range reg.SubexpNames() {
		if i == 0 || name == "" {
			continue
		}
		captures[name] = match[i]
	}
	return captures, true
}

// ParseStringByteSize return big size
func ParseStringByteSize(key string, defValue ...*big.Int) *big.Int {
	groups, matched := FindStringSubmatchMap(key, bitReg)
	if !matched {
		if len(defValue) == 0 {
			return nil
		}
		return defValue[0]
	}
	i, _ := ToInt64(groups["value"])

	switch groups["unit"] {
	case "b", "byte", "bytes":
		return (&big.Int{}).Mul(big.NewInt(i), _Byte)
	case "kb", "kilobyte", "kilobytes":
		return (&big.Int{}).Mul(big.NewInt(i), _KByte)
	case "mb", "megabyte", "megabytes":
		return (&big.Int{}).Mul(big.NewInt(i), _MByte)
	case "gb", "gigabyte", "gigabytes":
		return (&big.Int{}).Mul(big.NewInt(i), _GByte)
	case "tb", "terabyte", "terabytes":
		return (&big.Int{}).Mul(big.NewInt(i), _TByte)
	case "pb", "petabyte", "petabytes":
		return (&big.Int{}).Mul(big.NewInt(i), _PByte)
	case "eb", "exabyte", "exabytes":
		return (&big.Int{}).Mul(big.NewInt(i), _EByte)
	case "zb", "zettabyte", "zettabytes":
		return (&big.Int{}).Mul(big.NewInt(i), _ZByte)
	case "yb", "yottabyte", "yottabytes":
		return (&big.Int{}).Mul(big.NewInt(i), _YByte)
	case "k", "ki", "kib", "kibibyte", "kibibytes":
		return (&big.Int{}).Mul(big.NewInt(i), _KiByte)
	case "m", "mi", "mib", "mebibyte", "mebibytes":
		return (&big.Int{}).Mul(big.NewInt(i), _MiByte)
	case "g", "gi", "gib", "gibibyte", "gibibytes":
		return (&big.Int{}).Mul(big.NewInt(i), _GiByte)
	case "t", "ti", "tib", "tebibyte", "tebibytes":
		return (&big.Int{}).Mul(big.NewInt(i), _TiByte)
	case "p", "pi", "pib", "pebibyte", "pebibytes":
		return (&big.Int{}).Mul(big.NewInt(i), _PiByte)
	case "e", "ei", "eib", "exbibyte", "exbibytes":
		return (&big.Int{}).Mul(big.NewInt(i), _EiByte)
	case "z", "zi", "zib", "zebibyte", "zebibytes":
		return (&big.Int{}).Mul(big.NewInt(i), _ZiByte)
	case "y", "yi", "yib", "yobibyte", "yobibytes":
		return (&big.Int{}).Mul(big.NewInt(i), _YiByte)
	default:
		if len(defValue) == 0 {
			return nil
		}
		return defValue[0]
	}
}

// ParseStringTime return time.Duration
func ParseStringTime(s string, defValue ...time.Duration) time.Duration {
	groups, matched := FindStringSubmatchMap(s, timeReg)

	if !matched {
		return defaultDuration(defValue...)
	}

	i, err := ToInt64(groups["value"])
	if err != nil {
		return defaultDuration(defValue...)
	}

	switch groups["unit"] {
	case "nanoseconds", "nanosecond", "nanos", "nano", "ns":
		return checkedDuration(time.Nanosecond, i, defValue...)
	case "microseconds", "microsecond", "micros", "micro", "us":
		return checkedDuration(time.Microsecond, i, defValue...)
	case "milliseconds", "millisecond", "millis", "milli", "ms":
		return checkedDuration(time.Millisecond, i, defValue...)
	case "seconds", "second", "s":
		return checkedDuration(time.Second, i, defValue...)
	case "minutes", "minute", "m":
		return checkedDuration(time.Minute, i, defValue...)
	case "hours", "hour", "h":
		return checkedDuration(time.Hour, i, defValue...)
	case "days", "day", "d":
		return checkedDuration(24*time.Hour, i, defValue...)
	case "weeks", "week", "w":
		return checkedDuration(7*24*time.Hour, i, defValue...)
	case "years", "year", "y":
		return checkedDuration(365*24*time.Hour, i, defValue...)
	default:
		return defaultDuration(defValue...)
	}
}

func defaultDuration(defValue ...time.Duration) time.Duration {
	if len(defValue) == 0 {
		return 0
	}
	return defValue[0]
}

func checkedDuration(unit time.Duration, value int64, defValue ...time.Duration) time.Duration {
	const maxDuration = time.Duration(1<<63 - 1)

	if value < 0 || value > int64(maxDuration)/int64(unit) {
		return defaultDuration(defValue...)
	}
	return unit * time.Duration(value)
}
