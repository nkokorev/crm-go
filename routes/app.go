package routes

import (
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/controllers"
	"github.com/nkokorev/crm-go/middleware"
	"net/http"
)

/**
* [App UI-API] - group of routes for working http://app.ratuscrm.com
*
* Context(r): issuerAccount = RatusCRM (*models.Account)
* Context(r): targetAccount (or account) = loaded Account (*models.Account)

* В контексте rApp accountId = 1 (RatusCRM)
* В контексте rAppAuthUser, accountId = 1 (RatusCRM)
* В контексте rAppAuthFull accountId/userId в зависимости от аккаунта
 */
//var AppRoutes = func(rApp, rAppAuthUser, rAppAuthFull *mux.Router) {
var AppRoutes = func(rApp *mux.Router) {

	// 1. Create Auth User or User & Account routes
	rApp.Use(middleware.CheckAppUiApiStatus, middleware.AddContextMainAccount)

	rAuthUser := rApp.PathPrefix("").Subrouter()
	rAuthFull := rApp.PathPrefix("").Subrouter()
	rAppAccId := rApp.PathPrefix("").Subrouter()

	// 2. Add middleware
	rAuthUser.Use(middleware.CheckAppUiApiStatus, middleware.AddContextMainAccount, middleware.JwtCheckUserAuthentication)
	rAuthFull.Use(middleware.CheckAppUiApiStatus, middleware.AddContextMainAccount, middleware.JwtCheckFullAuthentication)

	// 3. Load system routes
	rApp.HandleFunc("/", controllers.CheckAppUiApi).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)
	rApp.HandleFunc("/app/settings", controllers.CrmGetSettings).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)

	// 4. Load check authentication routes
	rAuthUser.HandleFunc("/app/auth/check/user", controllers.AuthenticationJWTCheck).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/app/auth/check/account", controllers.AuthenticationJWTCheck).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/app/auth/check/full", controllers.AuthenticationJWTCheck).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)

	// 5. Load authentication routes in App (id of issuerAccount = 1 alw)
	rApp.HandleFunc("/users/auth/username", controllers.UserAuthByUsername).Methods(http.MethodPost, http.MethodOptions)
	rApp.HandleFunc("/users/auth/email", controllers.UserAuthByEmail).Methods(http.MethodPost, http.MethodOptions)
	rApp.HandleFunc("/users/auth/phone", controllers.UserAuthByPhone).Methods(http.MethodPost, http.MethodOptions)

	// 6. Load sign-in routes (account get from hash id)
	rAppAccId.HandleFunc("/accounts/{accountId:[0-9]+}/users", controllers.UserRegistration).Methods(http.MethodPost, http.MethodOptions)
	rAppAccId.HandleFunc("/accounts/{accountId:[0-9]+}/users/auth/username", controllers.UserAuthByUsername).Methods(http.MethodPost, http.MethodOptions)

	// 7. Test marketing: test Email...
	rAuthFull.HandleFunc("/accounts/marketing/test-email", controllers.SendEmailMessage).
		Methods(http.MethodPost, http.MethodOptions)

	// 8. For User auth accounts
	rAppAuthUserAccId := rAuthUser.PathPrefix("").Subrouter()
	rAppAuthUserAccId.Use(middleware.ContextMuxVarIssuerAccountId) // получаем ID issuerAccountId

	// 9.
	rAuthUser.HandleFunc("/accounts/{accountId:[0-9]+}/auth", controllers.AccountAuthUser).Methods(http.MethodPost, http.MethodOptions)
	rAuthUser.HandleFunc("/users/accounts", controllers.UserGetAccounts).Methods(http.MethodGet, http.MethodOptions)

}
