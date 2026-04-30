package config

// DeepCopy returns a deep copy of value. For map[string]any, map[any]any, and
// []any, nested maps and slices are copied recursively. Scalars and other
// values are returned unchanged (including shared references for types not
// handled above).
func DeepCopy(value any) any {
	if value == nil {
		return nil
	}
	switch x := value.(type) {
	case map[string]any:
		dst := make(map[string]any, len(x))
		for k, v := range x {
			dst[k] = DeepCopy(v)
		}
		return dst
	case map[any]any:
		dst := make(map[any]any, len(x))
		for k, v := range x {
			dst[DeepCopy(k)] = DeepCopy(v)
		}
		return dst
	case []any:
		dst := make([]any, len(x))
		for i := range x {
			dst[i] = DeepCopy(x[i])
		}
		return dst
	default:
		return x
	}
}
