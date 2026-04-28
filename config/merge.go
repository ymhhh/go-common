package config

// deepMerge merges src into dst (in-place) and returns dst.
// When both sides are objects (map[string]any), it merges recursively.
// Otherwise src overwrites dst.
func deepMerge(dst, src map[string]any) map[string]any {
	if dst == nil {
		dst = map[string]any{}
	}
	for k, sv := range src {
		if sm, ok := sv.(map[string]any); ok {
			if dv, ok := dst[k]; ok {
				if dm, ok := dv.(map[string]any); ok {
					dst[k] = deepMerge(dm, sm)
					continue
				}
			}
			dst[k] = deepMerge(map[string]any{}, sm)
			continue
		}
		dst[k] = sv
	}
	return dst
}
