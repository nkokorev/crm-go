package routes

import (
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/controllers"
	"net/http"
)

/**
* [App UI-API] - группа роутов для работы основного приложения app.ratuscrm.com
*
* В контексте issuerAccountID = 1 (всегда!).

* В контексте всегда есть {account_id}. Для базовых запросов он равен 1 (RatusCRM)
* В контексте rApp accountId = 1 (RatusCRM)
* В контексте rAppAuthUser, accountId = 1 (RatusCRM)
* В контексте rAppAuthFull accountId/userId в зависимости от аккаунта
 */
var AppRoutes = func (rApp, rAppAuthUser, rAppAuthFull *mux.Router) {

	// загружаем базовые настройки системы
	rApp.HandleFunc("/", controllers.CheckAppUiApi).Methods(http.MethodGet, http.MethodPost, http.MethodOptions) // +


	// AccountId = 1
	rApp.HandleFunc("/users", controllers.UserRegistration).Methods(http.MethodPost, http.MethodOptions)
	rApp.HandleFunc("/users/auth/username", controllers.UserAuthByUsername).Methods(http.MethodPost, http.MethodOptions)
	rApp.HandleFunc("/users/auth/email", controllers.UserAuthByEmail).Methods(http.MethodPost, http.MethodOptions)
	rApp.HandleFunc("/users/auth/phone", controllers.UserAuthByPhone).Methods(http.MethodPost, http.MethodOptions)


	//rApp.HandleFunc("/auth/user", controllers.AuthenticationJWTCheck).Methods(http.MethodGet, http.MethodOptions)
	//rApp.HandleFunc("/auth/account", controllers.AuthenticationJWTCheck).Methods(http.MethodGet, http.MethodOptions)
	//rApp.HandleFunc("/auth", controllers.AuthenticationJWTCheck).Methods(http.MethodGet, http.MethodOptions)

}