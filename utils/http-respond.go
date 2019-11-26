package utils

import (
	//"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"github.com/json-iterator/go"
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

// todo дописать нормально
func MessageError(message string, err error) (map[interface{}]interface{}) {
	e := Error{}
	errors.As(err, &e)

	mp := map[interface{}]interface{} {
		"status" : false,
		"message" : message,
		"errors" : e.GetErrors(),
	}
	fmt.Println(mp)
	return map[interface{}]interface{} {
		"status" : false,
		"message" : message,
		"errors" : e.GetErrors(),
	}
}
