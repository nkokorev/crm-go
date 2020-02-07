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
var UiApiRoutes = func (rUiApi, rUiApiAuthFull *mux.Router) {

	// загружаем базовые настройки системы
	rUiApi.HandleFunc("/", controllers.CrmGetSettings).Methods(http.MethodGet, http.MethodOptions)


}