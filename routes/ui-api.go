package routes

import (
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/controllers"
	"net/http"
)

/**
* [App UI-API] - группа роутов для работы основного приложения app.ratuscrm.com
*
* Оба роутера монтируются в точку /accounts/{accountId} имеют в контексте account & accountId
* rUiApi - маршруты без проверки JWT
* rUiApiAuth - маршрут с проверкой JWT, а также на совпадение {accountId} с accountId указаном в токене
 */
var UiApiRoutes = func (rFree, rUiApiAuthFull *mux.Router) {

	// загружаем базовые настройки системы
	rFree.HandleFunc("/", controllers.CheckUiApi).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)

	rFree.HandleFunc("/users", controllers.UserRegistration).Methods(http.MethodPost, http.MethodOptions)
	rFree.HandleFunc("/users/auth/username", controllers.UserAuthByUsername).Methods(http.MethodPost, http.MethodOptions)
	rFree.HandleFunc("/users/auth/email", controllers.UserAuthByEmail).Methods(http.MethodPost, http.MethodOptions)
	rFree.HandleFunc("/users/auth/phone", controllers.UserAuthByPhone).Methods(http.MethodPost, http.MethodOptions)


}