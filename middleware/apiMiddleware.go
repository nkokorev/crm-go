package middleware

import (
	"context"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
	"strings"
)

// Проверяет общий статус API интерфейса для ВСЕХ аккаунтов
func ApiEnabled(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Подгружаем настройки CRM
		crmSettings, err := models.GetCrmSettings()

		if err != nil {
			u.Respond(w, u.MessageError(nil, "Server is unavailable")) // что это?)
			return
		}

		// Проверяем статус API для всех клиентов
		if !crmSettings.ApiEnabled {
			u.Respond(w, u.MessageError(nil, crmSettings.ApiDisabledMessage))
			return
		}

		next.ServeHTTP(w, r) //proceed in the middleware chain!
	})
}

// проверяет валидность 32-символного api-ключа
func BearerAuthentication(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Получаем заголовок авторизации
		tokenHeader := r.Header.Get("Authorization") //Grab the token from the header

		if tokenHeader == "" { //Token is missing, returns with error code 403 Unauthorized
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, u.Message(false, "Header Authorization not found"))
			return
		}

		splitted := strings.Split(tokenHeader, " ") //The token normally comes in format `Bearer {token-body}`, we check if the retrieved token matched this requirement
		if len(splitted) != 2 {
			resp := u.Message(false, "Invalid/Malformed auth token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, resp)
			return
		}

		// Собираем вторую часть строки "Bearer <...>"
		token := splitted[1] //Grab the token part, what we are truly interested in

		// ищем ApiKey с указанным токеном
		key, err := models.GetApiKey(token)
		if err != nil {
			resp := u.Message(false, "Invalid/Malformed auth token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, resp)
			return
		}

		if !key.Status {
			resp := u.Message(false, "Auth token is disabled")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, resp)
			return
		}

		// ищем аккаунт, связанный с apiKey
		account, err := models.GetAccount(key.AccountID)
		if err != nil || account == nil {
			resp := u.Message(false, "Account not found, plz refresh auth token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, resp)
			return
		}

		// Проверяем статус включения API в аккаунте
		if !account.ApiEnabled {
			resp := u.Message(false, "Sorry, API is disabled for current account")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, resp)
			return
		}

		// имеет смысл, т.к. все запросы быдут связаны с аккаунтом
		r = r.WithContext(context.WithValue(r.Context(), "author", "api"))
		r = r.WithContext(context.WithValue(r.Context(), "account", account))
		r = r.WithContext(context.WithValue(r.Context(), "accountID", account.ID))

		next.ServeHTTP(w, r) //proceed in the middleware chain!
	})
}
