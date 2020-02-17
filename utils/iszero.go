package utils

import (
	"errors"
	"reflect"
)

func IsZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		z := true
		for i := 0; i < v.Len(); i++ {
			z = z && IsZero(v.Index(i))
		}
		return z
	case reflect.Struct:
		z := true
		for i := 0; i < v.NumField(); i++ {
			z = z && IsZero(v.Field(i))
		}
		return z
	}
	// Compare other types directly:
	z := reflect.Zero(v.Type())
	return v.Interface() == z.Interface()
}

func CheckNotNullFields(input interface{}, enum []string) error {
	var e Error

	val := reflect.ValueOf(input)
	//val := reflect.ValueOf(input).Elem()

	// if its a pointer, resolve its value
	if val.Kind() == reflect.Ptr {
		val = reflect.Indirect(val)
	}

	// should double check we now have a struct (could still be anything)
	if val.Kind() != reflect.Struct {
		return errors.New("unexpected type")
	}

	for _,v := range enum {

		check := false

		for i:=0; i < val.NumField();i++ {

			name := ToLowerCamel(val.Type().Field(i).Name)

			// Проверка наличия поля
			if name == v && !IsZero(reflect.ValueOf(input).Elem().Field(i)){
				check = true
				continue
			}
		}

		if check == false {
			e.AddErrors(v, "Необходимо заполнить поле")
		}

	}

	if e.HasErrors() {
		e.Message = "Проверьте правильность заполнения формы"
		return e
	} else {
		return nil
	}
}