package config

import (
	"fmt"
	"strings"
)

func splitPath(path string) ([]string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("config: empty path")
	}
	parts := strings.Split(path, ".")
	for _, p := range parts {
		if p == "" {
			return nil, fmt.Errorf("config: invalid path %q", path)
		}
	}
	return parts, nil
}

func getPath(root map[string]any, path string) (any, bool) {
	parts, err := splitPath(path)
	if err != nil {
		return nil, false
	}
	var cur any = root
	for _, p := range parts {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[p]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}

func setPath(root map[string]any, path string, value any) error {
	parts, err := splitPath(path)
	if err != nil {
		return err
	}

	cur := root
	for i := 0; i < len(parts)-1; i++ {
		p := parts[i]
		next, ok := cur[p]
		if !ok || next == nil {
			nm := map[string]any{}
			cur[p] = nm
			cur = nm
			continue
		}
		nm, ok := next.(map[string]any)
		if !ok {
			// replace non-map with map to continue path
			nm = map[string]any{}
			cur[p] = nm
		}
		cur = nm
	}
	cur[parts[len(parts)-1]] = value
	return nil
}
