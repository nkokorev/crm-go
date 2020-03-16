package middleware

import (
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
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
