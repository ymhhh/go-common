package config

import "reflect"

// DeepCopy returns a deep copy of value. Maps and slices are copied
// recursively; scalars and other values are returned unchanged.
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
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Map:
			return deepCopyMap(rv).Interface()
		case reflect.Slice:
			return deepCopySlice(rv).Interface()
		}
		return x
	}
}

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
	for _, key := range src.MapKeys() {
		dst.SetMapIndex(
			deepCopyAs(key, src.Type().Key()),
			deepCopyAs(src.MapIndex(key), src.Type().Elem()),
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
		dst.Index(i).Set(deepCopyAs(src.Index(i), src.Type().Elem()))
	}
	return dst
}

func deepCopyAs(src reflect.Value, target reflect.Type) reflect.Value {
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
	if src.Type().AssignableTo(target) {
		return src
	}
	if src.Type().ConvertibleTo(target) {
		return src.Convert(target)
	}
	return reflect.Zero(target)
}
