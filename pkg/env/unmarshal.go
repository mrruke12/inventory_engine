package env

import (
	"fmt"
	"reflect"
)

// Unmarshals values from environment into the given structure according
// to the tags of its fields (or field names if no tags are present).
//
// Supports int, string data types; returns an error for others.
// The dest must be a pointer to a structure.
//
//	struct Config {
//		Host string `env:"DB_HOST"`
//		Port int	`env:"DB_PORT"`
//	}
//	cfg := &Config{}
//	err := UnmarshalStruct(cfg)
func UnmarshalStruct(dest any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("unepxected error in UnmarshalConfig: %v", r)
		}
	}()

	pt := reflect.TypeOf(dest)

	if pt.Kind() != reflect.Ptr || pt.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("%w: got %v", ErrNotPointerToStruct, pt)
	}

	t := pt.Elem()
	v := reflect.ValueOf(dest).Elem()

	for i := range t.NumField() {
		ft := t.Field(i)
		fv := v.Field(i)

		tag, exists := ft.Tag.Lookup("env")

		var key string

		if exists {
			key = tag
		} else {
			key = ft.Name
		}

		var newVal any

		switch fv.Kind() {
		case reflect.Int:
			newVal, err = RequireInt(key)
		case reflect.String:
			newVal, err = Require(key)
		default:
			return ErrUnsupportedFieldType
		}

		if err != nil {
			return err
		}

		fv.Set(reflect.ValueOf(newVal))
	}

	return nil
}
