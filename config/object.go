package config

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

func decodeToObject(v any, out any) error {
	if out == nil {
		return fmt.Errorf("config: out is nil")
	}

	// Common shorthand: scalar string to a structured config.
	// - If out implements encoding.TextUnmarshaler, use it.
	// - Otherwise, if out is a pointer to struct and has a well-known string field
	//   (Type/Kind/Name/Driver/Parser), assign it.
	if s, ok := v.(string); ok {
		if tu, ok := out.(encoding.TextUnmarshaler); ok {
			if err := tu.UnmarshalText([]byte(s)); err != nil {
				return fmt.Errorf("config: unmarshal text: %w", err)
			}
			return nil
		}

		rv := reflect.ValueOf(out)
		if rv.Kind() == reflect.Pointer && !rv.IsNil() && rv.Elem().Kind() == reflect.Struct {
			st := rv.Elem()
			for _, fieldName := range []string{"Type", "Kind", "Name", "Driver", "Parser"} {
				f := st.FieldByName(fieldName)
				if f.IsValid() && f.CanSet() && f.Kind() == reflect.String {
					f.SetString(s)
					return nil
				}
			}
		}
	}

	// Prefer JSON round-trip for consistent behavior with struct tags (`json:"..."`).
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("config: marshal: %w", err)
	}
	if err := json.Unmarshal(b, out); err == nil {
		return nil
	}

	// Fallback to YAML to better support yaml tags.
	yb, _ := yaml.Marshal(v)
	if err := yaml.Unmarshal(yb, out); err != nil {
		return fmt.Errorf("config: unmarshal to object: %w", err)
	}
	return nil
}
