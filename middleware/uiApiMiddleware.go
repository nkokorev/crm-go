package middleware

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
	"strings"
)

// Проверяет общий статус API интерфейса для ВСЕХ аккаунтов
func CheckApiStatus(next http.Handler) http.Handler {

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

// Проверяет статус App UI-API (для GUI RatusCRM)
func CheckAppUiApiStatus(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("AppUiApiEnabled")
		// Подгружаем настройки CRM
		crmSettings, err := models.GetCrmSettings()

		if err != nil {
			u.Respond(w, u.MessageError(nil, "Server is unavailable")) // что это?)
			return
		}

		// Проверяем статус UI-API для главного приложения (APP/GUI)
		if !crmSettings.AppUiApiEnabled {
			u.Respond(w, u.MessageError(nil, crmSettings.AppUiApiDisabledMessage))
			return
		}

		next.ServeHTTP(w, r) //proceed in the middleware chain!
	})
}

// Проверяет общий статус UI-API интерфейса для ВСЕХ аккаунтов
func CheckUiApiStatus(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Подгружаем настройки CRM
		crmSettings, err := models.GetCrmSettings()

		if err != nil {
			u.Respond(w, u.MessageError(nil, "Сервер не может обработать запрос")) // что это?)
			return
		}

		// Проверяем статус UI-API для всех клиентов
		if !crmSettings.UiApiEnabled {
			u.Respond(w, u.MessageError(nil, crmSettings.UiApiDisabledMessage))
			return
		}

		next.ServeHTTP(w, r)
	})

}

// Вставляет в контекст accountId из hashId
func ContextMuxVarAccount(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		accountHashId := mux.Vars(r)["accountHashId"] // защищаемся от парсинга / спама

		if len(accountHashId) != 12 {
			u.Respond(w, u.MessageError(nil, "The hashID length must be 12 symbols"))
			return
		}

		account,err := models.GetAccountByHash(accountHashId)
		if err != nil {
			u.Respond(w, u.MessageError(nil, "An account with the specified hash ID was not found"))
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "auth", "ui/api"))

		ctx1 := context.WithValue(r.Context(), "account", account)
		r = r.WithContext(ctx1)

		ctx2 := context.WithValue(r.Context(), "accountId", account.ID)
		r = r.WithContext(ctx2)

		next.ServeHTTP(w, r)
	})

}

// Вставляет в контекст accountId из JWT
func ContextMainAccount(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// тут могла быть специальная функция, берущая данные из памяти (чекая по времени)
		account,err := models.GetAccount(1) // RatusCRM
		if err != nil {
			u.Respond(w, u.MessageError(nil, "An account with the specified hash ID was not found"))
			return
		}

		// For future
		r = r.WithContext(context.WithValue(r.Context(), "auth", "app ui/api"))

		ctx1 := context.WithValue(r.Context(), "account", account)
		r = r.WithContext(ctx1)

		ctx2 := context.WithValue(r.Context(), "accountId", account.ID)
		r = r.WithContext(ctx2)

		next.ServeHTTP(w, r)
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

		if !key.Enabled {
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
		r = r.WithContext(context.WithValue(r.Context(), "auth", "api"))
		r = r.WithContext(context.WithValue(r.Context(), "account", account))
		r = r.WithContext(context.WithValue(r.Context(), "accountId", account.ID))

		next.ServeHTTP(w, r) //proceed in the middleware chain!
	})
}

