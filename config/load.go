package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

	ext := strings.ToLower(filepath.Ext(path))

	var root map[string]any
	var incFromLines []string
	var cleaned []byte
	switch ext {
	case ".json":
		raw, err = stripJSONComments(raw)
		if err != nil {
			return nil, fmt.Errorf("config: parse %s: %w", path, err)
		}
		incFromLines, cleaned = processIncludes(raw)
		root, err = parseJSON(cleaned)
	case ".yaml", ".yml":
		// YAML treats "#include ..." as a comment, but we strip it for consistency.
		incFromLines, cleaned = processIncludes(raw)
		root, err = parseYAML(cleaned)
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
	var err error
	raw, err = stripJSONComments(raw)
	if err != nil {
		return nil, err
	}

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
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	merged := map[string]any{}
	seenDoc := false

	for {
		var v any
		if err := dec.Decode(&v); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if v == nil {
			continue
		}
		m, ok := normalize(v).(map[string]any)
		if !ok {
			return nil, fmt.Errorf("config: yaml root must be map/object")
		}
		seenDoc = true
		merged = deepMerge(merged, m)
	}

	if !seenDoc {
		return nil, fmt.Errorf("config: yaml root must be map/object")
	}
	return merged, nil
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

// processIncludes extracts column-zero #include directives and returns the
// cleaned content in a single pass. Indented "#include ..." text is data (for
// example YAML block scalar content), not a directive.
func processIncludes(raw []byte) (incs []string, cleaned []byte) {
	lines := bytes.Split(raw, []byte{'\n'})
	filtered := make([][]byte, 0, len(lines))
	for _, ln := range lines {
		s := string(ln)
		if strings.HasPrefix(s, "#include ") {
			p := strings.TrimSpace(strings.TrimPrefix(s, "#include "))
			if p != "" {
				incs = append(incs, p)
			}
			continue
		}
		filtered = append(filtered, ln)
	}
	return incs, bytes.Join(filtered, []byte{'\n'})
}
