package utils

import (
	"encoding/json"
	"net/http"
)

func Message(status bool, message string) (map[string]interface{}) {
	return map[string]interface{} {"status" : status, "message" : message}
}

func Respond(w http.ResponseWriter, data map[string] interface{})  {
	w.Header().Add("Content-Type", "application/json")
	//w.Header().Set("Access-Control-Allow-Origin", "*")
	//fmt.Println("Respond")
	json.NewEncoder(w).Encode(data)
}

func MessageWithErrors(message string, errors map[string]string) (map[string]interface{}) {
	return map[string]interface{} {"status" : false, "message" : message, "errors" : errors}
}
