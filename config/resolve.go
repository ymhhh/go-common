package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var refRe = regexp.MustCompile(`\$\{([^}]+)\}`)

type refLookup func(ref string) (any, bool)

func lookupEnv(key string) (string, bool) {
	v, ok := os.LookupEnv(key)
	return v, ok
}

func resolveAll(v any, lookup refLookup) error {
	visiting := map[string]bool{}
	return resolveAny(v, lookup, visiting)
}

func resolveAny(v any, lookup refLookup, visiting map[string]bool) error {
	switch x := v.(type) {
	case map[string]any:
		for k, vv := range x {
			// resolve value first
			if err := resolveAny(vv, lookup, visiting); err != nil {
				return err
			}
			x[k] = vv
			// if value is string, resolve placeholders and possibly replace type
			if s, ok := x[k].(string); ok {
				rv, err := resolveString(s, lookup, visiting)
				if err != nil {
					return err
				}
				x[k] = rv
			}
		}
		return nil
	case []any:
		for i := range x {
			if err := resolveAny(x[i], lookup, visiting); err != nil {
				return err
			}
			if s, ok := x[i].(string); ok {
				rv, err := resolveString(s, lookup, visiting)
				if err != nil {
					return err
				}
				x[i] = rv
			}
		}
		return nil
	case string:
		_, err := resolveString(x, lookup, visiting)
		return err
	default:
		return nil
	}
}

func resolveString(s string, lookup refLookup, visiting map[string]bool) (any, error) {
	matches := refRe.FindAllStringSubmatchIndex(s, -1)
	if len(matches) == 0 {
		return s, nil
	}

	// If the whole string is exactly "${...}", replace with referenced value (preserve type).
	if len(matches) == 1 && matches[0][0] == 0 && matches[0][1] == len(s) {
		ref := strings.TrimSpace(s[matches[0][2]:matches[0][3]])
		return resolveRefValue(ref, lookup, visiting)
	}

	// Otherwise, do string interpolation. Non-string ref values are formatted.
	var b strings.Builder
	last := 0
	for _, m := range matches {
		b.WriteString(s[last:m[0]])
		ref := strings.TrimSpace(s[m[2]:m[3]])
		rv, err := resolveRefValue(ref, lookup, visiting)
		if err != nil {
			return nil, err
		}
		switch x := rv.(type) {
		case nil:
			// empty
		case string:
			b.WriteString(x)
		default:
			b.WriteString(fmt.Sprint(x))
		}
		last = m[1]
	}
	b.WriteString(s[last:])
	return b.String(), nil
}

func resolveRefValue(ref string, lookup refLookup, visiting map[string]bool) (any, error) {
	if visiting[ref] {
		return nil, fmt.Errorf("config: reference cycle: %s", ref)
	}
	visiting[ref] = true
	defer delete(visiting, ref)

	v, ok := lookup(ref)
	if !ok {
		// For env-style placeholders, allow missing to resolve to empty string.
		// This is a pragmatic default for config templates.
		if _, envOk := os.LookupEnv(ref); !envOk {
			return "", nil
		}
	}

	// If the referenced value itself contains references, resolve it recursively.
	switch x := v.(type) {
	case string:
		return resolveString(x, lookup, visiting)
	case map[string]any, []any:
		if err := resolveAny(v, lookup, visiting); err != nil {
			return nil, err
		}
		return v, nil
	default:
		return v, nil
	}
}
