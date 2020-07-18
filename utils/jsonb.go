package utils

import (
	"encoding/json"
	"github.com/jinzhu/gorm/dialects/postgres"
)


func FixJSONB_String(input map[string]interface{}, keys []string) map[string]interface{} {

	if len(keys) < 1 || len(input) < 1 {
		return input
	}

	// Делаем апдейт по ключам
	for _,key := range keys {

		// получаем список array для каждого ключа
		arrI, ok := input[key].([]interface{})
		if !ok || arrI == nil { continue }

		// Собираем массив из []string
		var arrSTR = make([]string, 0)
		for i := range arrI {
			s, ok := arrI[i].(string)
			if !ok { continue }
			arrSTR = append(arrSTR, s)
		}

		// Преобразуем в JSON
		rawJSON, err := json.Marshal(arrSTR)
		if err != nil {
			rawJSON = json.RawMessage(`{}`)
		}
		
		input[key] = postgres.Jsonb{RawMessage: rawJSON}
	}
	

	return input
}
func FixJSONB_Uint(input map[string]interface{}, keys []string) map[string]interface{} {

	if len(keys) < 1 || len(input) < 1 {
		return input
	}

	// Делаем апдейт по ключам
	for _,key := range keys {

		// получаем список array для каждого ключа
		arrI, ok := input[key].([]interface{})
		if !ok || arrI == nil { continue }

		// Собираем массив из []string
		var arr = make([]uint, 0)
		for i := range arrI {
			s, ok := arrI[i].(float64)
			if !ok { continue }
			arr = append(arr, uint(s))
		}

		// Преобразуем в JSON
		rawJSON, err := json.Marshal(arr)
		if err != nil {
			rawJSON = json.RawMessage(`{}`)
		}

		input[key] = postgres.Jsonb{RawMessage: rawJSON}
	}


	return input
}

func NullArrIfNotFound() {

	/*rawJSON, err := json.Marshal([]string{})
	if err != nil {
		rawJSON = json.RawMessage(`{}`)
	}*/
}
