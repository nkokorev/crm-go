package utils

import (
	"encoding/json"
	"github.com/jinzhu/gorm/dialects/postgres"
	"gorm.io/datatypes"
)

// For function Update()
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

// For function Update()
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

func FixJSONB_MapString(input map[string]interface{}, keys []string) map[string]interface{} {

	if len(keys) < 1 || len(input) < 1 {
		return input
	}

	// Делаем апдейт по ключам
	for _,key := range keys {

		// 1
		arrMapString, ok := input[key].(map[string]interface{})
		if !ok || arrMapString == nil {
			continue
		}

		// 2. Преобразуем в JSON
		rawJSON, err := json.Marshal(arrMapString)
		if err != nil {
			rawJSON = json.RawMessage(`{}`)
		}

		input[key] = postgres.Jsonb{RawMessage: rawJSON}
	}


	return input
}

func ParseJSONBToString(jsonb datatypes.JSON) []string {

	var data = make([]string,0)

	b, err := jsonb.MarshalJSON()
	if err != nil {
		return data
	}
	
	if err := json.Unmarshal(b, &data); err != nil {
		return data
	}
	
	return data
}


// func ParseJSONBToMapString(jsonb postgres.Jsonb) map[string]interface{} {
func ParseJSONBToMapString(jsonb datatypes.JSON) map[string]interface{} {

	var data map[string]interface{}
	b, err := jsonb.MarshalJSON()
	if err != nil {
		return data
	}

	if err := json.Unmarshal(b, &data); err != nil {
		return data
	}

	return data
}


// удаляет все переменные которые с '_name'
func FixInputHiddenVars(input *map[string]interface{}) {
	for key := range *input {
		if string(key[0]) == "_" {
			delete(*input, key)
		}
	}
	delete(*input, "preloads")
}

// удаляет все переменные которые с '_name'
func FixInputDataTimeVars(input map[string]interface{}, keys []string) map[string]interface{} {
	for _,key := range keys {
		sR, ok := input[key].(float64)
		delete(input, key)
		if ok {
			var _time int = int(sR)
			input[key] = _time

			// _time, err := time.Parse(time.RFC3339, strconv.Itoa(int(sR)))
			/*_time, err := time.Parse(time.RFC3339, strconv.Itoa(int(sR)))
			if err == nil {
				input[key] = _time
			}*/
		}
	}
	return input
}