package utils

import (
	"errors"
	"github.com/json-iterator/go"
	"net/http"
)

func Message(status bool, message string) (map[interface{}]interface{}) {
	return map[interface{}]interface{} {"status" : status, "message" : message}
}

func Respond(w http.ResponseWriter, data map[interface{}] interface{}) {
	w.Header().Add("Content-Type", "application/json")
	//w.Header().Set("Access-Control-Allow-Origin", "*")
	//fmt.Println("Respond")

	jsoniter.NewEncoder(w).Encode(data)
}

func MessageWithErrors(message string, errors map[interface{}]interface{}) (map[interface{}]interface{}) {
	return map[interface{}]interface{} {"status" : false, "message" : message, "errors" : errors}
}

// Готовит сообщение с ошибкой
func MessageError(err error) (map[interface{}]interface{}) {

	e := Error{}
	resp := map[interface{}]interface{}{}
	resp["status"] = false

	if errors.As(err, &e) {
		resp["message"] = e.Message

		if len(e.Errors) > 0 {
			resp["errors"] = e.GetErrors()
		}
	} else {
		resp["message"] = err.Error()
	}

	return resp
}
