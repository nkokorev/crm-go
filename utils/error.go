package utils

import (
	"fmt"
)

type Error struct {
	Message string `json:"message"`
	Errors map[interface{}]interface{} `json:"message"`
}

// Пиздец важная функция
func (e Error) Error() string {
	return fmt.Sprintf("%v: %v", e.Message, e.Errors)
}

func (error Error) HasErrors() (status bool) {

	status = false

	if len(error.Message) > 0 || len(error.Errors) > 0 {
		status = true
	}
	return
}

func (e *Error) AddErrors(key interface{}, value interface{}) {
	if e.Errors == nil {
		//e.Errors = make(map[]interface{})
		e.Errors = make(map[interface{}]interface{})
	}
	e.Errors[key] = value
}

func (e *Error) GetErrors() map[interface{}]interface{} {
	if e.Errors == nil {
		e.Errors = make(map[interface{}]interface{})
	}
	return e.Errors
}

func (e *Error) GetError(key interface{}) interface{} {
	if e.Errors == nil {
		e.Errors = make(map[interface{}]interface{})
	}
	return e.Errors[key]
}


