package config

import "reflect"

// GetTag returns the value of a struct field tag for the given field and tag key.
// Usage: GetTag(MyConfig{}, "FieldName", "validate")
func GetTag(v any, field, tag string) (string, bool) {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return "", false
	}
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return "", false
	}
	if f, ok := rv.Type().FieldByName(field); ok {
		val, ok2 := f.Tag.Lookup(tag)
		return val, ok2
	}
	return "", false
}

// GetTags returns a map of fieldName -> tagValue for all exported fields for a given tag key.
// Only exported fields are included. Missing tags are skipped.
func GetTags(v any, tag string) map[string]string {
	out := map[string]string{}
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return out
	}
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return out
	}
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.PkgPath != "" { // unexported
			continue
		}
		if val, ok := f.Tag.Lookup(tag); ok {
			out[f.Name] = val
		}
	}
	return out
}
