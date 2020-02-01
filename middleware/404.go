package middleware

import (
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func NotFoundHandler () http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		//w.WriteHeader(http.StatusOK)
		u.Respond(w, u.Message(false, "Запрашиваемый объект не найден"))
		return
	});

}