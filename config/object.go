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

	if s, ok := v.(string); ok {
		assigned, err := assignStringShorthand(s, out)
		if assigned || err != nil {
			return err
		}
	}

	return roundTripDecode(v, out)
}

func assignStringShorthand(s string, out any) (bool, error) {
	if tu, ok := out.(encoding.TextUnmarshaler); ok {
		if err := tu.UnmarshalText([]byte(s)); err != nil {
			return true, fmt.Errorf("config: unmarshal text: %w", err)
		}
		return true, nil
	}

	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Pointer || rv.IsNil() || rv.Elem().Kind() != reflect.Struct {
		return false, nil
	}

	st := rv.Elem()
	for _, fieldName := range []string{"Type", "Kind", "Name", "Driver", "Parser"} {
		f := st.FieldByName(fieldName)
		if f.IsValid() && f.CanSet() && f.Kind() == reflect.String {
			f.SetString(s)
			return true, nil
		}
	}
	return false, nil
}

func roundTripDecode(v, out any) error {
	// Prefer JSON round-trip for consistent behavior with struct tags (`json:"..."`).
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("config: marshal: %w", err)
	}
	if err := json.Unmarshal(b, out); err == nil {
		return nil
	}

	// Fallback to YAML to better support yaml tags.
	yb, yerr := yaml.Marshal(v)
	if yerr != nil {
		return fmt.Errorf("config: marshal: %w", yerr)
	}
	if err := yaml.Unmarshal(yb, out); err != nil {
		return fmt.Errorf("config: unmarshal to object: %w", err)
	}
	return nil
}
