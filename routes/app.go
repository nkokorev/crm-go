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


	// 2. Add middleware
	rAuthUser.Use(middleware.CheckAppUiApiStatus, middleware.AddContextMainAccount, middleware.JwtCheckUserAuthentication)
	rAuthFull.Use(middleware.CheckAppUiApiStatus, middleware.AddContextMainAccount, middleware.JwtCheckFullAuthentication)

	// 3. Load system settings
	r.HandleFunc("/", controllers.CheckAppUiApi).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)
	r.HandleFunc("/app/settings", controllers.GetCRMSettings).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)

	// 4. App Authentication: /app/auth...
	rAuthUser.HandleFunc("/app/auth/check/user", controllers.AuthenticationJWTCheck).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/app/auth/check/account", controllers.AuthenticationJWTCheck).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/app/auth/check/full", controllers.AuthenticationJWTCheck).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)

	// 5. App Authentication user: Load authentication routes in App (id of issuerAccount = 1)
	// Тут базовая авторизая пользователя (не в аккаунте, а в issuer account'e)
	r.HandleFunc("/app/auth/username", controllers.UserAuthByUsername).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/app/auth/email", controllers.UserAuthByEmail).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/app/auth/phone", controllers.UserAuthByPhone).Methods(http.MethodPost, http.MethodOptions)

	// 6. Load sign-in routes (account get from hash id)
	// rAuthFull.HandleFunc("/accounts", controllers.AccountGetProfile).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{hashId}", controllers.AccountGetProfile).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{hashId}", controllers.AccountUpdate).Methods(http.MethodPatch, http.MethodOptions)
	// -- USERS --
	// Запрос ниже может иметь много параметров (диапазон выборки, число пользователей)
	rAuthFull.HandleFunc("/accounts/{hashId}/users", controllers.AccountUserList).Methods(http.MethodGet, http.MethodOptions)

	// rAppAccId.HandleFunc("/accounts/{hashId}/users", controllers.UserRegistration).Methods(http.MethodPost, http.MethodOptions)
	// rAppAccId.HandleFunc("/accounts/{hashId}/users/auth/username", controllers.UserAuthByUsername).Methods(http.MethodPost, http.MethodOptions)

	rAuthUser.HandleFunc("/accounts/{hashId}/auth", controllers.AccountAuthUser).Methods(http.MethodPost, http.MethodOptions)

	// ######## Uses #########
	// -- CRUD --
	rAuthUser.HandleFunc("/users/{hashId}/accounts", controllers.UserAccountsGet).Methods(http.MethodGet, http.MethodOptions)
	//rAuthUser.HandleFunc("/users/accounts", controllers.UserAccountsGet).Methods(http.MethodGet, http.MethodOptions)

	// ### ApiKeys ###
	// -- CRUD --
	rAuthFull.HandleFunc("/api-keys", controllers.ApiKeyGetCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/api-keys", controllers.ApiKeyGetList).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/api-keys/{id}", controllers.ApiKeyGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/api-keys/{id}", controllers.ApiKeyGetUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/api-keys/{id}", controllers.ApiKeyGetDelete).Methods(http.MethodDelete, http.MethodOptions)

	// ### EmailMarketing ###
	// ---CRUD---
	rAuthFull.HandleFunc("/domains", controllers.DomainsGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/email-templates", controllers.EmailTemplatesCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/email-templates", controllers.EmailTemplatesGetList).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/email-templates/{id}", controllers.EmailTemplateGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/email-templates/{id}", controllers.EmailTemplatesUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/email-templates/{id}", controllers.EmailTemplatesDelete).Methods(http.MethodDelete, http.MethodOptions)



	// ---ACCOUNT---
	rAuthFull.HandleFunc("/email-templates/{emailTemplateHashId}/send/user", controllers.EmailTemplateSendToUser).Methods(http.MethodPost, http.MethodOptions)

	// ### STORAGE CRUD ####
	rAuthFull.HandleFunc("/storage", controllers.StorageCreateFile).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/storage/{hashId}", controllers.StorageGetFileByHashId).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/storage", controllers.StorageGetList).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/storage/{hashId}", controllers.StorageUpdateFile).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/storage/{hashId}", controllers.StorageDeleteFile).Methods(http.MethodDelete, http.MethodOptions)

}
