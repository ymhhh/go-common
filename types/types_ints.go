package types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// ToInt64 parse value to int64
func ToInt64(value any) (int64, error) {
	if value == nil {
		return 0, nil
	}
	var val string
	switch reflect.TypeOf(value).Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int, reflect.Int32, reflect.Int64:
		val = fmt.Sprintf("%d", value)
	case reflect.String:
		switch reflect.TypeOf(value).String() {
		case "json.Number":
			return value.(json.Number).Int64()
		default:
			val = value.(string)
		}
	default:
		return 0, fmt.Errorf("type is valid: %s", reflect.TypeOf(value).String())
	}

	return strconv.ParseInt(val, 10, 64)
}

// ToInt parse value to int
func ToInt(value any) (int, error) {
	if value == nil {
		return 0, nil
	}

	var val string
	switch reflect.TypeOf(value).Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int, reflect.Int32, reflect.Int64:
		val = fmt.Sprintf("%d", value)
	case reflect.String:
		switch reflect.TypeOf(value).String() {
		case "json.Number":
			val = value.(json.Number).String()
		default:
			val = value.(string)
		}
	default:
		return 0, fmt.Errorf("type is valid: %s", reflect.TypeOf(value).String())
	}

	return strconv.Atoi(val)
}
