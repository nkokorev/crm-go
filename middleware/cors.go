package middleware

import (
	"net/http"
)

func CorsAccessControl(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// 1. Проверка источника
		// allowAuth := []string{"http://app.ratuscrm.me","http://localhost:8090","http://app.ratuscrm.com","https://app.ratuscrm.com"} //List of endpoints that doesn't require auth

		// 2. Frontend request host
		// requestHost := r.Header.Get("Origin") //current request path

		// w.Header().Add("Content-Type", "application/json;charset=UTF-8")
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Max-Age", "86400") // max 600
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,OPTIONS,POST,PUT,PATCH,DELETE")
		// w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,User-Agent")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization,Accept-Language,Cache-Control,Content-Type,Content-Length,Accept,Origin,X-Requested-With,Access-Control-Request-Headers,Access-Control-Request-Method,Access-Control-Allow-Credentials,Host, Origin, User-Agent, Referer")

		// 3. Выставляем cors-политику, если источник в разрешенных хостах
		/*for _, value := range allowAuth {

			if value == requestHost {
				w.Header().Add("Content-Type", "application/json;charset=UTF-8")
				//w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,OPTIONS,POST,PUT,PATCH")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization,Accept-Language,Cache-Control,Content-Type,Content-Length,Accept,Origin,X-Requested-With,Access-Control-Request-Headers,Access-Control-Request-Method,Access-Control-Allow-Credentials,Host, Origin, User-Agent, Referer")
				w.Header().Set("Access-Control-Max-Age", "600") // max for FireFox, Chrome max 600
				break
			}
		}*/

		// 3. Не передаем запрос дальше, если мето OPTIONS
		if r.Method == http.MethodOptions {
			return
		}

		next.ServeHTTP(w, r) //proceed in the middleware chain!
	})
}



