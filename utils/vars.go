package utils

import (
	"reflect"
)

func UINTp(x uint) *uint {
	return &x
}
func STRp(s string) *string {
	return &s
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