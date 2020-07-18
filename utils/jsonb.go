package utils

import (
	"encoding/json"
	"github.com/jinzhu/gorm/dialects/postgres"
)


func FixJSONB(input map[string]interface{}, keys []string) map[string]interface{} {

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

func NullArrIfNotFound() {

	/*rawJSON, err := json.Marshal([]string{})
	if err != nil {
		rawJSON = json.RawMessage(`{}`)
	}*/
}
