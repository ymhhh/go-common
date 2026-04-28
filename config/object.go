package config

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

func decodeToObject(v any, out any) error {
	if out == nil {
		return fmt.Errorf("config: out is nil")
	}

	// Prefer JSON round-trip for consistent behavior with struct tags (`json:"..."`).
	b, err := json.Marshal(v)
	if err == nil {
		if err := json.Unmarshal(b, out); err == nil {
			return nil
		}
	}

	// Fallback to YAML to better support yaml tags or special types if needed.
	yb, yerr := yaml.Marshal(v)
	if yerr != nil {
		if err != nil {
			return fmt.Errorf("config: marshal failed (json=%v, yaml=%v)", err, yerr)
		}
		return fmt.Errorf("config: marshal yaml: %w", yerr)
	}
	if err := yaml.Unmarshal(yb, out); err != nil {
		return fmt.Errorf("config: unmarshal to object: %w", err)
	}
	return nil
}
