package utils

import (
	"errors"
	"github.com/json-iterator/go"
	"net/http"
)

func Message(status bool, message string) (map[string]interface{}) {
	return map[string]interface{} {"status" : status, "message" : message}
}

func Respond(w http.ResponseWriter, data map[string] interface{}) {
	w.Header().Add("Content-Type", "application/json")
	//w.Header().Set("Access-Control-Allow-Origin", "*")
	//fmt.Println("Respond")

	jsoniter.NewEncoder(w).Encode(data)
}

func MessageWithErrors(message string, errors map[string]interface{}) (map[string]interface{}) {
	return map[string]interface{} {"status" : false, "message" : message, "errors" : errors}
}

// Возвращает сообщение с ошибкой. Если err можно привести к u.Error, если нет - смотрит, если ли запасной параметр.
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
