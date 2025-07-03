package pipz

import (
	"reflect"
)

// getTypeName returns the string representation of a type
// This is pipz's own implementation to avoid circular dependencies
func getTypeName[T any]() string {
	var zero T
	t := reflect.TypeOf(zero)
	if t == nil {
		return "nil"
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		if t == nil {
			return "nil"
		}
	}
	if t.Name() != "" {
		return t.Name()
	}
	return t.String()
}