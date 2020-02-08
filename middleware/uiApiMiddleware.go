package middleware

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

// Проверяет общий статус UI-API для APP/GUI
func AppUiApiEnabled(next http.Handler) http.Handler {

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
func UiApiEnabled(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Подгружаем настройки CRM
		crmSettings, err := models.GetCrmSettings()

		if err != nil {
			u.Respond(w, u.MessageError(nil, "Server is unavailable")) // что это?)
			return
		}

		// Проверяем статус UI-API для всех клиентов
		if !crmSettings.UiApiEnabled {
			u.Respond(w, u.MessageError(nil, crmSettings.UiApiDisabledMessage))
			return
		}

		next.ServeHTTP(w, r) //proceed in the middleware chain!
	})
}

// Вставляет в контекст accountId
func CheckAccountId(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		accountId := mux.Vars(r)["accountId"]

		fmt.Println("accountId: ", accountId)

		next.ServeHTTP(w, r)
	})

}

func CheckUiApiEnabled(next http.Handler) http.Handler {

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


// app.ratuscrm.com/ui-api/.... Bearer token <>
func UiApiJWTAuthentication(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// 1. Подгрузим файл настроек, он тут будет кстати
		crmSettings, err := models.GetCrmSettings()
		if err != nil {
			u.Respond(w, u.MessageError(nil, "Сервер не может обработать запрос")) // что это?)
			return
		}

		// если Public UI-API погашен
		if !crmSettings.UiApiEnabled {
			u.Respond(w, u.MessageError(nil, crmSettings.UiApiDisabledMessage)) // что это?)
			return
		}

		// Пробуем получить account_id



		/*ctx1 := context.WithValue(r.Context(), "user_id", tk.UserId)
		r = r.WithContext(ctx1)

		ctx2 := context.WithValue(r.Context(), "account_id", tk.AccountId)
		r = r.WithContext(ctx2)*/

		next.ServeHTTP(w, r) //proceed in the middleware chain!
	});
}

// All request {/accounts/{account_id}/...}
