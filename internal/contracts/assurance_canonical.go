package contracts

import (
	"reflect"
	"strings"
)

func canonicalize(v any) any {
	switch t := v.(type) {
	case map[string]any:
		return canonicalizeMap(t)
	case []any:
		return canonicalizeSlice(t)
	case string:
		return t
	default:
		return canonicalizeByKind(indirect(reflect.ValueOf(v)), v)
	}
}

func canonicalizeMap(m map[string]any) any {
	out := make(map[string]any, len(m))
	for key, val := range m {
		out[key] = canonicalize(val)
	}
	return out
}

func canonicalizeSlice(items []any) any {
	out := make([]any, len(items))
	for i := range items {
		out[i] = canonicalize(items[i])
	}
	return out
}

func canonicalizeByKind(rv reflect.Value, fallback any) any {
	if !rv.IsValid() {
		return nil
	}
	switch rv.Kind() {
	case reflect.Map:
		return canonicalizeReflectMap(rv)
	case reflect.Array, reflect.Slice:
		return canonicalizeReflectSlice(rv)
	case reflect.Struct:
		return canonicalizeStruct(rv)
	case reflect.String:
		return rv.String()
	default:
		return fallback
	}
}

func canonicalizeReflectMap(rv reflect.Value) any {
	if rv.Type().Key().Kind() != reflect.String {
		return rv.Interface()
	}
	m := make(map[string]any, rv.Len())
	for _, key := range rv.MapKeys() {
		m[key.String()] = canonicalize(rv.MapIndex(key).Interface())
	}
	return m
}

func canonicalizeReflectSlice(rv reflect.Value) any {
	length := rv.Len()
	arr := make([]any, length)
	for i := 0; i < length; i++ {
		arr[i] = canonicalize(rv.Index(i).Interface())
	}
	return arr
}

func canonicalizeStruct(rv reflect.Value) any {
	result := make(map[string]any)
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if field.PkgPath != "" && !field.Anonymous {
			continue
		}
		value := rv.Field(i)
		fieldName, include := canonicalFieldName(field)
		if !include || !value.CanInterface() || shouldSkipJSONField(field, value) {
			continue
		}
		if field.Anonymous {
			mergeAnonymousField(result, value)
			continue
		}
		result[fieldName] = canonicalize(value.Interface())
	}
	return result
}

func canonicalFieldName(field reflect.StructField) (string, bool) {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return "", false
	}
	if tag == "" {
		return field.Name, true
	}
	parts := strings.Split(tag, ",")
	if parts[0] == "" {
		return field.Name, true
	}
	return parts[0], true
}

func shouldSkipJSONField(field reflect.StructField, value reflect.Value) bool {
	tag := field.Tag.Get("json")
	if tag == "" {
		return false
	}
	parts := strings.Split(tag, ",")
	for _, option := range parts[1:] {
		if option == "omitempty" && value.IsZero() {
			return true
		}
	}
	return false
}

func mergeAnonymousField(result map[string]any, value reflect.Value) {
	nested := canonicalize(value.Interface())
	nestedMap, ok := nested.(map[string]any)
	if !ok {
		return
	}
	for key, val := range nestedMap {
		result[key] = val
	}
}

func indirect(rv reflect.Value) reflect.Value {
	for rv.Kind() == reflect.Interface || rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return rv
		}
		rv = rv.Elem()
	}
	return rv
}
