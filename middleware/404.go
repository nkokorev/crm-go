package middleware

import (
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func NotFoundHandler () http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		u.Respond(w, u.Message(false, "Не найден вызываемый URL"))
		return
	});

}

func NotFoundMethod () http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		u.Respond(w, u.Message(false, "HTTP-метод указан не верно"))
		return
	});

}