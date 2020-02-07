package routes

import (
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/controllers"
	"net/http"
)

/**
* [App UI-API] - группа роутов для работы основного приложения app.ratuscrm.com
 */
var AppRoutes = func (rApp, rAppAuthUser, rAppAuthFull *mux.Router) {

	// загружаем базовые настройки системы
	rApp.HandleFunc("/settings", controllers.CrmGetSettings).Methods(http.MethodGet, http.MethodOptions)


	//rApp.HandleFunc("/auth/user", controllers.AuthenticationJWTCheck).Methods(http.MethodGet, http.MethodOptions)
	//rApp.HandleFunc("/auth/account", controllers.AuthenticationJWTCheck).Methods(http.MethodGet, http.MethodOptions)
	//rApp.HandleFunc("/auth", controllers.AuthenticationJWTCheck).Methods(http.MethodGet, http.MethodOptions)

}