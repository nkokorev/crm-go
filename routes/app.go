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
var AppRoutes = func(r *mux.Router) {

	// 1. Create more rotes [User] or [Full] (User & Account)
	// r.Use(middleware.CheckAppUiApiStatus, middleware.AddContextMainAccount)

	rAuthUser := r.PathPrefix("").Subrouter()
	rAuthFull := r.PathPrefix("").Subrouter()
	rAppAccId := r.PathPrefix("").Subrouter()


	// 2. Add middleware
	rAuthUser.Use(middleware.CheckAppUiApiStatus, middleware.AddContextMainAccount, middleware.JwtCheckUserAuthentication)
	rAuthFull.Use(middleware.CheckAppUiApiStatus, middleware.AddContextMainAccount, middleware.JwtCheckFullAuthentication)
	rAppAccId.Use(middleware.CheckAppUiApiStatus, middleware.AddContextMainAccount, middleware.JwtCheckUserAuthentication, middleware.ContextMuxVarIssuerAccountId)

	// 3. Load system settings
	r.HandleFunc("/", controllers.CheckAppUiApi).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)
	r.HandleFunc("/app/settings", controllers.GetCRMSettings).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)

	// 4. App Authentication: /app/auth...
	rAuthUser.HandleFunc("/app/auth/check/user", controllers.AuthenticationJWTCheck).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/app/auth/check/account", controllers.AuthenticationJWTCheck).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/app/auth/check/full", controllers.AuthenticationJWTCheck).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)

	// 5. App Authentication user: Load authentication routes in App (id of issuerAccount = 1)
	r.HandleFunc("/app/auth/username", controllers.UserAuthByUsername).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/app/auth/email", controllers.UserAuthByEmail).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/app/auth/phone", controllers.UserAuthByPhone).Methods(http.MethodPost, http.MethodOptions)

	// 6. Load sign-in routes (account get from hash id)
	rAppAccId.HandleFunc("/accounts/{accountId:[0-9]+}/users", controllers.UserRegistration).Methods(http.MethodPost, http.MethodOptions)
	rAppAccId.HandleFunc("/accounts/{accountId:[0-9]+}/users/auth/username", controllers.UserAuthByUsername).Methods(http.MethodPost, http.MethodOptions)

	// ### EmailMarketing ###
	rAuthFull.HandleFunc("/accounts/domains", controllers.DomainsGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/email-templates/{emailTemplateHashId}", controllers.EmailTemplateGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/email-templates", controllers.EmailTemplatesGetList).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/email-templates", controllers.EmailTemplatesDelete).Methods(http.MethodDelete, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/email-templates", controllers.EmailTemplatesCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/email-templates", controllers.EmailTemplatesUpdate).Methods(http.MethodPatch, http.MethodOptions)

	// ### Orders ###
	rAppAccId.HandleFunc("/accounts/{accountId:[0-9]+}/orders", controllers.GetOrders).Methods(http.MethodGet, http.MethodOptions)

	// 7. Test marketing: test Email...
	rAuthFull.HandleFunc("/accounts/marketing/test-email", controllers.SendEmailMessage).
		Methods(http.MethodPost, http.MethodOptions)

	// 8. For User authentication accounts
	rAppAuthUserAccId := rAuthUser.PathPrefix("").Subrouter()
	rAppAuthUserAccId.Use(middleware.ContextMuxVarIssuerAccountId) // получаем ID issuerAccountId

	// 9. Support HZ function
	rAppAccId.HandleFunc("/accounts/{accountId:[0-9]+}/auth", controllers.AccountAuthUser).Methods(http.MethodPost, http.MethodOptions)
	rAuthUser.HandleFunc("/users/accounts", controllers.UserGetAccounts).Methods(http.MethodGet, http.MethodOptions)

}
