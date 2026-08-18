package config

import "reflect"

// DeepCopy returns a deep copy of value. Maps (any key/value types) and
// slices (any element type) are copied recursively. Scalars and other
// immutable values are returned unchanged.
func DeepCopy(value any) any {
	if value == nil {
		return nil
	}

	// Fast path for the most common config types (no reflection).
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
	}

	// Reflection path for typed maps and slices (e.g. map[string]string,
	// []int, etc.).
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Map:
		return deepCopyMap(rv).Interface()
	case reflect.Slice:
		return deepCopySlice(rv).Interface()
	case reflect.Array:
		return deepCopyArray(rv).Interface()
	case reflect.Pointer:
		if rv.IsNil() {
			return nil
		}
		// If pointer points to a mutable composite, copy the pointed-to value
		// and return a pointer to the copy.
		elem := rv.Elem()
		switch elem.Kind() {
		case reflect.Map, reflect.Slice, reflect.Array:
			copied := DeepCopy(elem.Interface())
			if copied == nil {
				return reflect.Zero(rv.Type()).Interface()
			}
			ptr := reflect.New(elem.Type())
			ptr.Elem().Set(reflect.ValueOf(copied))
			return ptr.Interface()
		}
		return value
	default:
		return value
	}
}

// isMutableComposite reports whether value is a map, slice, or array that
// could share mutable state if not deep-copied.
func isMutableComposite(value any) bool {
	if value == nil {
		return false
	}
	switch reflect.ValueOf(value).Kind() {
	case reflect.Map, reflect.Slice:
		return true
	default:
		return false
	}
}

func deepCopyMap(src reflect.Value) reflect.Value {
	if src.IsNil() {
		return reflect.Zero(src.Type())
	}
	dst := reflect.MakeMapWithSize(src.Type(), src.Len())
	iter := src.MapRange()
	for iter.Next() {
		dst.SetMapIndex(
			deepCopyElem(iter.Key(), src.Type().Key()),
			deepCopyElem(iter.Value(), src.Type().Elem()),
		)
	}
	return dst
}

func deepCopySlice(src reflect.Value) reflect.Value {
	if src.IsNil() {
		return reflect.Zero(src.Type())
	}
	dst := reflect.MakeSlice(src.Type(), src.Len(), src.Len())
	for i := range src.Len() {
		dst.Index(i).Set(deepCopyElem(src.Index(i), src.Type().Elem()))
	}
	return dst
}

func deepCopyArray(src reflect.Value) reflect.Value {
	dst := reflect.New(src.Type()).Elem()
	for i := range src.Len() {
		dst.Index(i).Set(deepCopyElem(src.Index(i), src.Type().Elem()))
	}
	return dst
}

// deepCopyElem deep-copies src.Interface() and ensures the result is
// assignable to target. If the deep copy produces an incompatible type
// (should not happen for valid config data), a zero value is returned
// rather than falling back to a shared reference.
func deepCopyElem(src reflect.Value, target reflect.Type) reflect.Value {
	if !src.IsValid() {
		return reflect.Zero(target)
	}
	copied := DeepCopy(src.Interface())
	if copied == nil {
		return reflect.Zero(target)
	}
	cv := reflect.ValueOf(copied)
	if cv.Type().AssignableTo(target) {
		return cv
	}
	if cv.Type().ConvertibleTo(target) {
		return cv.Convert(target)
	}
	// Should not be reached for types encountered in config trees.
	return reflect.Zero(target)
}
