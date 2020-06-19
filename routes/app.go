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
	rAuthFull.HandleFunc("/accounts/{accountHashId}", controllers.AccountGetProfile).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}", controllers.AccountUpdate).Methods(http.MethodPatch, http.MethodOptions)
	// -- USERS --
	// Запрос ниже может иметь много параметров (диапазон выборки, число пользователей)
	// rAuthFull.HandleFunc("/accounts/{accountHashId}/users", controllers.CreateUser).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/users", controllers.GetUserList).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/users/{userHashId}", controllers.RemoveUserFromAccount).Methods(http.MethodDelete, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/users/{userHashId}", controllers.UpdateUserData).Methods(http.MethodPatch, http.MethodOptions)

	// -- ROLES --
	// Запрос ниже может иметь много параметров (диапазон выборки, число пользователей)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/roles", controllers.RoleList).Methods(http.MethodGet, http.MethodOptions)

	rAuthUser.HandleFunc("/accounts/{accountHashId}/auth", controllers.AccountAuthUser).Methods(http.MethodPost, http.MethodOptions)

	// ######## Uses #########
	// -- CRUD --
	rAuthUser.HandleFunc("/users/{hashId}/app/auth/username", controllers.UserAccountsGet).Methods(http.MethodGet, http.MethodOptions)
	//rAuthUser.HandleFunc("/users/accounts", controllers.UserAccountsGet).Methods(http.MethodGet, http.MethodOptions)

	// ### ApiKeys ###
	// -- CRUD --
	rAuthFull.HandleFunc("/api-keys", controllers.ApiKeyCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/api-keys", controllers.ApiKeyGetList).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/api-keys/{id}", controllers.ApiKeyGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/api-keys/{id}", controllers.ApiKeyUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/api-keys/{id}", controllers.ApiKeyDelete).Methods(http.MethodDelete, http.MethodOptions)

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
	rAuthFull.HandleFunc("/storage", controllers.StorageGetListPagination).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/storage/{id}", controllers.StorageGetFile).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/storage/{id}", controllers.StorageUpdateFile).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/storage/{id}", controllers.StorageDeleteFile).Methods(http.MethodDelete, http.MethodOptions)

	// ### SHOP ####
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops", controllers.ShopCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops", controllers.ShopListGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}", controllers.ShopUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}", controllers.ShopDelete).Methods(http.MethodDelete, http.MethodOptions)

	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}/product-groups", controllers.ProductGroupCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}/product-groups", controllers.ProductGroupListPaginationByShopGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}/product-groups/{productGroupId:[0-9]+}", controllers.ProductGroupByShopGet).Methods(http.MethodGet, http.MethodOptions)
	//rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}/product-groups", controllers.ProductGroupListGet).Methods(http.MethodGet, http.MethodOptions)
	//rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/product-groups", controllers.ProductGroupListGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}/product-groups/{groupId:[0-9]+}", controllers.ProductGroupUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}/product-groups/{groupId:[0-9]+}", controllers.ProductGroupDelete).Methods(http.MethodDelete, http.MethodOptions)

	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/product-cards", controllers.ProductCardCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}/product-cards", controllers.ProductCardByShopCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}/product-cards/{productCardId:[0-9]+}", controllers.ProductCardByShopGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}/product-cards", controllers.ProductCardListPaginationByShopGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/product-cards/{cardId:[0-9]+}", controllers.ProductCardUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/product-cards/{cardId:[0-9]+}", controllers.ProductCardDelete).Methods(http.MethodDelete, http.MethodOptions)

	// todo: byShop - сейчас косвенная принадлежность товаров к магазину не отслеживается
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/products", controllers.ProductCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/products/{productId:[0-9]+}", controllers.ProductGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/products", controllers.ProductListPaginationGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/products/{productId:[0-9]+}", controllers.ProductUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/products/{productId:[0-9]+}", controllers.ProductDelete).Methods(http.MethodDelete, http.MethodOptions)

}
