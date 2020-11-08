package utils

import (
	"reflect"
	"time"
)

func UINTp(x uint) *uint {
	return &x
}
func INTp(x int) *int {
	return &x
}
func ParseUINTp(x *uint) uint {
	if x == nil {
		return 0
	}
	return *x
}

func FL64p(x float64) *float64 {
	return &x
}
func STRp(s string) *string {
	return &s
}
func TimeP(t time.Time) *time.Time {
	return &t
}
func ConvertMapVarsToUINT(input *map[string]interface{}, keys []string) error {

	for _,key := range keys {
		if _vI, ok := (*input)[key]; ok && _vI != nil {
			switch reflect.TypeOf(_vI).String() {

			// case "float64":
			case "uint":
				(*input)[key] = _vI
			case "int":
				_vFInt, ok := _vI.(int)
				if !ok {
					return Error{Message: "Техническая ошибка в запросе", Errors: map[string]interface{}{key:"не удается разобрать значение"}}
				}
				(*input)[key] = uint(_vFInt)
			case "string":
				(*input)[key] = nil
				// fmt.Println("Char")
			default:
				_vF64, ok := _vI.(float64)
				if !ok {
					return Error{Message: "Техническая ошибка в запросе", Errors: map[string]interface{}{key:"не удается разобрать значение"}}
				}
				(*input)[key] = uint(_vF64)
			}
		}
	}
	return nil
}

func ConvertMapVarsToFloat64(input *map[string]interface{}, keys []string) error {

	for _,key := range keys {
		if _vI, ok := (*input)[key]; ok && _vI != nil {
			switch reflect.TypeOf(_vI).String() {

			// case "float64":
			case "uint":
				(*input)[key] = _vI
			case "int":
				_vFInt, ok := _vI.(int)
				if !ok {
					return Error{Message: "Техническая ошибка в запросе", Errors: map[string]interface{}{key:"не удается разобрать значение"}}
				}
				(*input)[key] = uint(_vFInt)
			default:
				_vF64, ok := _vI.(float64)
				if !ok {
					return Error{Message: "Техническая ошибка в запросе", Errors: map[string]interface{}{key:"не удается разобрать значение"}}
				}
				(*input)[key] = uint(_vF64)
			}
		}
	}
	return nil
}

// возвращает только разрешенные ключи
func FilterAllowedKeySTRArray(input []string, keys []string) []string {

	retArr := make([]string,0)
	for inK := range input {
		for key := range keys {
			if ToCamel(input[inK]) == ToCamel(keys[key]) {
				retArr = append(retArr, keys[key])
			}
		}
	}
	return retArr
}