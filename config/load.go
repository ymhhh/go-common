package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const includeKey = "#include"

func loadFile(path string, stack map[string]struct{}) (map[string]any, error) {
	if _, ok := stack[path]; ok {
		return nil, fmt.Errorf("config: include cycle detected: %s", path)
	}
	stack[path] = struct{}{}
	defer delete(stack, path)

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}

	incFromLines := parseIncludeLines(raw)
	ext := strings.ToLower(filepath.Ext(path))

	var root map[string]any
	switch ext {
	case ".json":
		root, err = parseJSON(stripIncludeLines(raw))
	case ".yaml", ".yml":
		// YAML treats "#include ..." as a comment, but we strip it anyway for consistency.
		root, err = parseYAML(stripIncludeLines(raw))
	default:
		return nil, fmt.Errorf("config: unsupported file type: %s", ext)
	}
	if err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}

	baseDir := filepath.Dir(path)

	incs := make([]string, 0, len(incFromLines))
	incs = append(incs, incFromLines...)
	if v, ok := root[includeKey]; ok {
		for _, s := range toStringSlice(v) {
			incs = append(incs, s)
		}
		delete(root, includeKey)
	}

	merged := map[string]any{}
	for _, inc := range incs {
		ip := inc
		if !filepath.IsAbs(ip) {
			ip = filepath.Join(baseDir, ip)
		}
		ip, err = filepath.Abs(ip)
		if err != nil {
			return nil, fmt.Errorf("config: abs include path: %w", err)
		}
		im, err := loadFile(ip, stack)
		if err != nil {
			return nil, err
		}
		merged = deepMerge(merged, im)
	}

	merged = deepMerge(merged, root) // current overrides included
	return merged, nil
}

func parseJSON(raw []byte) (map[string]any, error) {
	raw = bytes.TrimSpace(raw)
	raw = stripJSONComments(raw)

	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()

	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	m, ok := normalize(v).(map[string]any)
	if !ok {
		return nil, fmt.Errorf("config: json root must be object")
	}
	return m, nil
}

func parseYAML(raw []byte) (map[string]any, error) {
	var v any
	if err := yaml.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	m, ok := normalize(v).(map[string]any)
	if !ok {
		return nil, fmt.Errorf("config: yaml root must be map/object")
	}
	return m, nil
}

func normalize(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, vv := range x {
			out[k] = normalize(vv)
		}
		return out
	case map[any]any:
		out := make(map[string]any, len(x))
		for k, vv := range x {
			out[fmt.Sprint(k)] = normalize(vv)
		}
		return out
	case []any:
		out := make([]any, 0, len(x))
		for _, vv := range x {
			out = append(out, normalize(vv))
		}
		return out
	default:
		return x
	}
}

func toStringSlice(v any) []string {
	switch x := v.(type) {
	case string:
		x = strings.TrimSpace(x)
		if x == "" {
			return nil
		}
		return []string{x}
	case []any:
		out := make([]string, 0, len(x))
		for _, it := range x {
			if s, ok := it.(string); ok {
				s = strings.TrimSpace(s)
				if s != "" {
					out = append(out, s)
				}
			}
		}
		return out
	default:
		return nil
	}
}

// parseIncludeLines supports a simple preprocessor directive:
//
//	#include path/to/other.yaml
//
// Lines are parsed before JSON/YAML decoding. This works well with JSONC where
// #include can be used as a standalone line. In YAML, it is treated as a comment,
// but we still honor it at this pre-parse stage.
func parseIncludeLines(raw []byte) []string {
	lines := bytes.Split(raw, []byte{'\n'})
	var out []string
	for _, ln := range lines {
		s := strings.TrimSpace(string(ln))
		if strings.HasPrefix(s, "#include ") {
			p := strings.TrimSpace(strings.TrimPrefix(s, "#include "))
			if p != "" {
				out = append(out, p)
			}
		}
	}
	return out
}

func stripIncludeLines(raw []byte) []byte {
	lines := bytes.Split(raw, []byte{'\n'})
	out := make([][]byte, 0, len(lines))
	for _, ln := range lines {
		s := strings.TrimSpace(string(ln))
		if strings.HasPrefix(s, "#include ") {
			continue
		}
		out = append(out, ln)
	}
	return bytes.Join(out, []byte{'\n'})
}
