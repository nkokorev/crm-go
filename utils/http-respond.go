package utils

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	jsoniter "github.com/json-iterator/go"
	"net/http"
	"strconv"
)

func Message(status bool, message string) (map[string]interface{}) {
	return map[string]interface{} {"status" : status, "message" : message}
}

func Respond(w http.ResponseWriter, data map[string] interface{}) {
	//w.Header().Add("Content-Type", "application/json")
	/*w.Header().Add("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Max-Age", "86400") // max 600*/
	// w.Header().Set("Access-Control-Allow-Origin", "*")
	jsoniter.NewEncoder(w).Encode(data)
	// json.NewEncoder(w).Encode(data)
}

func MessageWithErrors(message string, errors map[string]interface{}) (map[string]interface{}) {
	return map[string]interface{} {"status" : false, "message" : message, "errors" : errors}
}

// Отправляет error если она из utils.Error или opt_msg[0]
func MessageError(err error, opt_msg... string) (map[string]interface{}) {

	e := Error{}
	resp := map[string]interface{}{}
	resp["status"] = false

	if errors.As(err, &e) {
		resp["message"] = e.Message

		if len(e.Errors) > 0 {
			resp["errors"] = e.GetErrors()
		}
	} else {
		// выводить системные ошибки - плохо

		if len(opt_msg) > 0 {
			resp["message"] = opt_msg[0]
		} else {
			resp["message"] = err.Error()
		}

	}

	return resp
}

func GetFromRequestUINT(r *http.Request, val_name string) (uint, error) {

	vars := mux.Vars(r)
	str_int_id := vars[val_name]

	id, err := strconv.ParseUint(str_int_id, 10, 32)
	return uint(id), err
}

func MapToRawJson(input map[string]interface{}) json.RawMessage {

	b, err := json.Marshal(input)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return b
}

func StringArrToRawJson(input []string) json.RawMessage {

	b, err := json.Marshal(input)
	if err != nil {
		// return json.RawMessage(`{}`)
		return json.RawMessage(`[]`)
	}
	return b
}

func UINTArrToRawJson(input []uint) json.RawMessage {

	b, err := json.Marshal(input)
	if err != nil {
		// return json.RawMessage(`{}`)
		return json.RawMessage(`[]`)
	}
	return b
}