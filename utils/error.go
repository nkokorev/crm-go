package utils

import "fmt"

type Error struct {
	Message string `json:"message"`
	Errors map[string]string `json:"message"`
}

func (error Error) HasErrors() (status bool) {

	status = false

	if len(error.Message) > 0 || len(error.Errors) > 0 {
		status = true
	}
	return
}

func (e *Error) AddErrors(key, val string) {
	if e.Errors == nil {
		e.Errors = make(map[string]string)
	}
	e.Errors[key] = val
}

func (e *Error) GetErrors() map[string]string {
	if e.Errors == nil {
		e.Errors = make(map[string]string)
	}
	return e.Errors
}

func (e *Error) GetError(key string) string {
	if e.Errors == nil {
		e.Errors = make(map[string]string)
	}
	return e.Errors[key]
}

func (e *Error) SetMsg(msg string) {
	e.Message = msg
}

func (e *Error) GetMsg() string {
	return e.Message
}

func (e *Error) GetResponse() (string, map[string]string) {
	return e.GetMsg(), e.GetErrors()
}

func (e *Error) Println()  {
	fmt.Println(e.GetMsg())
	for _, r := range e.GetErrors() {
		fmt.Println(r)
	}
}