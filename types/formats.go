package types

import (
	"math/big"
	"regexp"
	"strconv"
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
		return defaultByteSize(defValue...)
	}

	switch groups["unit"] {
	case "b", "byte", "bytes":
		return parseByteSizeValue(groups["value"], _Byte, defValue...)
	case "kb", "kilobyte", "kilobytes":
		return parseByteSizeValue(groups["value"], _KByte, defValue...)
	case "mb", "megabyte", "megabytes":
		return parseByteSizeValue(groups["value"], _MByte, defValue...)
	case "gb", "gigabyte", "gigabytes":
		return parseByteSizeValue(groups["value"], _GByte, defValue...)
	case "tb", "terabyte", "terabytes":
		return parseByteSizeValue(groups["value"], _TByte, defValue...)
	case "pb", "petabyte", "petabytes":
		return parseByteSizeValue(groups["value"], _PByte, defValue...)
	case "eb", "exabyte", "exabytes":
		return parseByteSizeValue(groups["value"], _EByte, defValue...)
	case "zb", "zettabyte", "zettabytes":
		return parseByteSizeValue(groups["value"], _ZByte, defValue...)
	case "yb", "yottabyte", "yottabytes":
		return parseByteSizeValue(groups["value"], _YByte, defValue...)
	case "k", "ki", "kib", "kibibyte", "kibibytes":
		return parseByteSizeValue(groups["value"], _KiByte, defValue...)
	case "m", "mi", "mib", "mebibyte", "mebibytes":
		return parseByteSizeValue(groups["value"], _MiByte, defValue...)
	case "g", "gi", "gib", "gibibyte", "gibibytes":
		return parseByteSizeValue(groups["value"], _GiByte, defValue...)
	case "t", "ti", "tib", "tebibyte", "tebibytes":
		return parseByteSizeValue(groups["value"], _TiByte, defValue...)
	case "p", "pi", "pib", "pebibyte", "pebibytes":
		return parseByteSizeValue(groups["value"], _PiByte, defValue...)
	case "e", "ei", "eib", "exbibyte", "exbibytes":
		return parseByteSizeValue(groups["value"], _EiByte, defValue...)
	case "z", "zi", "zib", "zebibyte", "zebibytes":
		return parseByteSizeValue(groups["value"], _ZiByte, defValue...)
	case "y", "yi", "yib", "yobibyte", "yobibytes":
		return parseByteSizeValue(groups["value"], _YiByte, defValue...)
	default:
		return defaultByteSize(defValue...)
	}
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
	return new(big.Int).Quo(r.Num(), r.Denom())
}

// ParseStringTime return time.Duration
func ParseStringTime(s string, defValue ...time.Duration) time.Duration {
	groups, matched := FindStringSubmatchMap(s, timeReg)

	if !matched {
		if len(defValue) == 0 {
			return 0
		}
		return defValue[0]
	}

	f, err := strconv.ParseFloat(groups["value"], 64)
	if err != nil {
		if len(defValue) == 0 {
			return 0
		}
		return defValue[0]
	}

	switch groups["unit"] {
	case "nanoseconds", "nanosecond", "nanos", "nano", "ns":
		return time.Duration(float64(time.Nanosecond) * f)
	case "microseconds", "microsecond", "micros", "micro", "us":
		return time.Duration(float64(time.Microsecond) * f)
	case "milliseconds", "millisecond", "millis", "milli", "ms":
		return time.Duration(float64(time.Millisecond) * f)
	case "seconds", "second", "s":
		return time.Duration(float64(time.Second) * f)
	case "minutes", "minute", "m":
		return time.Duration(float64(time.Minute) * f)
	case "hours", "hour", "h":
		return time.Duration(float64(time.Hour) * f)
	case "days", "day", "d":
		return time.Duration(float64(time.Hour*24) * f)
	case "weeks", "week", "w":
		return time.Duration(float64(time.Hour*24*7) * f)
	case "years", "year", "y":
		return time.Duration(float64(time.Hour*24*365) * f)
	default:
		if len(defValue) == 0 {
			return 0
		}
		return defValue[0]
	}
}
