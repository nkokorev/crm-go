package middleware

import (
	"context"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
	"strings"
)

// проверяет валидность 32-символного api-ключа
func BearerAuthentication(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenHeader := r.Header.Get("Authorization") //Grab the token from the header

		if tokenHeader == "" { //Token is missing, returns with error code 403 Unauthorized
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, u.Message(false, "Ошибка авторизации: отсутствует заголовок Authorization."))
			return
		}

		splitted := strings.Split(tokenHeader, " ") //The token normally comes in format `Bearer {token-body}`, we check if the retrieved token matched this requirement
		if len(splitted) != 2 {
			resp := u.Message(false, "Invalid/Malformed auth token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, resp)
			return
		}

		// Собираем вторую часть строки "Bearer kSDkfslfds390d2w...."
		token := splitted[1] //Grab the token part, what we are truly interested in

		// ищем ApiKey с указанным токеном
		//key := &models.ApiKey{Token:token}

		// спорно
		key, err := models.GetApiKey(token)
		if err != nil {
			resp := u.Message(false, "Invalid/Malformed auth token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, resp)
			return
		}

		if !key.Status {
			resp := u.Message(false, "Auth token is disabled.")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, resp)
			return
		}

		ctx1 := context.WithValue(r.Context(), "account", key.Account)
		r = r.WithContext(ctx1)

		ctx2 := context.WithValue(r.Context(), "accountID", key.Account.ID)
		r = r.WithContext(ctx2)

		next.ServeHTTP(w, r) //proceed in the middleware chain!
	})
}
