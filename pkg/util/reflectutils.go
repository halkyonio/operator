package util

import (
	"fmt"
	"reflect"
)

// SetNamedStringField sets the `fieldName` field to `newFieldValue` on the `toUpdate` struct. Note that the struct must be
// passed as a pointer so that it can actually be modified. Returns whether the given struct was changed as a result or if the
// operation raised an error
func SetNamedStringField(toUpdate interface{}, fieldName string, newFieldValue string) (changed bool, err error) {
	valueToUpdate := reflect.ValueOf(toUpdate)
	if valueToUpdate.Kind() == reflect.Ptr {
		valueToUpdate = valueToUpdate.Elem()
	} else {
		return false, fmt.Errorf("must pass a pointer so that you can actually set the value, see https://blog.golang.com/laws-of-reflection, law 3")
	}
	if len(fieldName) > 0 {
		field := valueToUpdate.FieldByName(fieldName)
		if field.IsValid() && field.CanSet() {
			if field.Kind() == reflect.String {
				if field.String() != newFieldValue {
					field.SetString(newFieldValue)
					changed = true
				}
			} else {
				return false, fmt.Errorf("'%s' is a %s field on %s, need a string", fieldName, field.Kind(), valueToUpdate.Type().Name())
			}
		} else {
			return false, fmt.Errorf("invalid '%s' field for %s", fieldName, valueToUpdate.Type().Name())
		}
	}
	return
}

// MustSetNamedStringField is similar to SetNamedStringField except that raised errors will panic instead
func MustSetNamedStringField(toUpdate interface{}, fieldName string, newFieldValue string) (changed bool) {
	changed, err := SetNamedStringField(toUpdate, fieldName, newFieldValue)
	if err != nil {
		panic(err)
	}
	return changed
}
