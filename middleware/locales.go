package middleware

import (
	"net/http"
)

// Обработка языка запроса, чтобы узнать на каком языке отвечать
var I18nMiddleware = func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//accept := r.Header.Get("Accept-Language")
		//t.SetAccept(accept)
		next.ServeHTTP(w, r) //proceed in the middleware chain!
	})
}
