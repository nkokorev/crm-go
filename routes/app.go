package routes

import (
	"github.com/gorilla/mux"
	// "github.com/nkokorev/crm-go/controllers"
	"github.com/nkokorev/crm-go/controllers/appCr"
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
	r.HandleFunc("/", appCr.CheckAppUiApi).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)
	r.HandleFunc("/app/settings", appCr.GetCRMSettings).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)

	// 4. App Authentication: /app/auth...
	rAuthUser.HandleFunc("/app/auth/check/user", appCr.AuthenticationJWTCheck).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/app/auth/check/account", appCr.AuthenticationJWTCheck).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/app/auth/check/full", appCr.AuthenticationJWTCheck).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)

	// 5. App Authentication user: Load authentication routes in App (id of issuerAccount = 1)
	// Тут базовая авторизая пользователя (не в аккаунте, а в issuer account'e)
	r.HandleFunc("/app/auth/username", appCr.UserAuthByUsername).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/app/auth/email", appCr.UserAuthByEmail).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/app/auth/phone", appCr.UserAuthByPhone).Methods(http.MethodPost, http.MethodOptions)

	// 6. Load sign-in routes (account get from hash id)
	// rAuthFull.HandleFunc("/accounts", controllers.AccountGetProfile).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}", appCr.AccountGetProfile).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}", appCr.AccountUpdate).Methods(http.MethodPatch, http.MethodOptions)

	// -- USERS --
	// Запрос ниже может иметь много параметров (диапазон выборки, число пользователей)
	// rAuthFull.HandleFunc("/accounts/{accountHashId}/users", controllers.CreateUser).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/users", appCr.GetUserList).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/users/{userHashId}", appCr.RemoveUserFromAccount).Methods(http.MethodDelete, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/users/{userHashId}", appCr.UpdateUserData).Methods(http.MethodPatch, http.MethodOptions)

	// -- ROLES --
	// Запрос ниже может иметь много параметров (диапазон выборки, число пользователей)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/roles", appCr.RoleList).Methods(http.MethodGet, http.MethodOptions)

	rAuthUser.HandleFunc("/accounts/{accountHashId}/auth", appCr.AccountAuthUser).Methods(http.MethodPost, http.MethodOptions)

	// ######## Uses #########
	// -- CRUD --
	rAuthUser.HandleFunc("/users/{hashId}/app/auth/username", appCr.UserAccountsGet).Methods(http.MethodGet, http.MethodOptions)
	//rAuthUser.HandleFunc("/users/accounts", appCr.UserAccountsGet).Methods(http.MethodGet, http.MethodOptions)

	// ### ApiKeys ###
	// -- CRUD --
	rAuthFull.HandleFunc("/accounts/{accountHashId}/api-keys", appCr.ApiKeyCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/api-keys", appCr.ApiKeyGetList).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/api-keys/{id}", appCr.ApiKeyGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/api-keys/{id}", appCr.ApiKeyUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/api-keys/{id}", appCr.ApiKeyDelete).Methods(http.MethodDelete, http.MethodOptions)

	// ### EmailMarketing ###
	// ---CRUD---
	rAuthFull.HandleFunc("/accounts/{accountHashId}/domains", appCr.DomainsGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-templates", appCr.EmailTemplatesCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-templates", appCr.EmailTemplatesGetList).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-templates/{id}", appCr.EmailTemplateGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-templates/{id}", appCr.EmailTemplatesUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-templates/{id}", appCr.EmailTemplatesDelete).Methods(http.MethodDelete, http.MethodOptions)
	
	// ---ACCOUNT---
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-templates/{emailTemplateHashId}/send/user", appCr.EmailTemplateSendToUser).Methods(http.MethodPost, http.MethodOptions)

	// ### STORAGE CRUD ####
	rAuthFull.HandleFunc("/accounts/{accountHashId}/storage", appCr.StorageCreateFile).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/storage", appCr.StorageGetListPagination).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/storage/{fileId}", appCr.StorageGetFile).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/storage/{fileId}", appCr.StorageUpdateFile).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/storage/{fileId}", appCr.StorageDeleteFile).Methods(http.MethodDelete, http.MethodOptions)

	// ### SHOP ####
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops", appCr.ShopCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops", appCr.ShopListGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}", appCr.ShopUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}", appCr.ShopDelete).Methods(http.MethodDelete, http.MethodOptions)

	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}/product-groups", appCr.ProductGroupCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}/product-groups", appCr.ProductGroupListPaginationByShopGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}/product-groups/{productGroupId:[0-9]+}", appCr.ProductGroupByShopGet).Methods(http.MethodGet, http.MethodOptions)
	//rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}/product-groups", controllers.ProductGroupListGet).Methods(http.MethodGet, http.MethodOptions)
	//rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/product-groups", controllers.ProductGroupListGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}/product-groups/{groupId:[0-9]+}", appCr.ProductGroupUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}/product-groups/{groupId:[0-9]+}", appCr.ProductGroupDelete).Methods(http.MethodDelete, http.MethodOptions)

	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/product-cards", appCr.ProductCardCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}/product-cards", appCr.ProductCardByShopCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}/product-cards/{productCardId:[0-9]+}", appCr.ProductCardByShopGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}/product-cards", appCr.ProductCardListPaginationByShopGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/product-cards/{cardId:[0-9]+}", appCr.ProductCardUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/product-cards/{cardId:[0-9]+}", appCr.ProductCardDelete).Methods(http.MethodDelete, http.MethodOptions)

	// todo: byShop - сейчас косвенная принадлежность товаров к магазину не отслеживается
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/products", appCr.ProductCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/products/{productId:[0-9]+}", appCr.ProductGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/products", appCr.ProductListPaginationGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/products/{productId:[0-9]+}", appCr.ProductUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/products/{productId:[0-9]+}", appCr.ProductDelete).Methods(http.MethodDelete, http.MethodOptions)

	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}/deliveries", appCr.DeliveryCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}/deliveries", appCr.DeliveryListPaginationGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}/deliveries/{deliveryId:[0-9]+}", appCr.DeliveryUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}/deliveries/{deliveryId:[0-9]+}", appCr.DeliveryDelete).Methods(http.MethodDelete, http.MethodOptions)

	// ### WebHooks ####
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-hooks", appCr.WebHookCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-hooks", appCr.WebHookListPaginationGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-hooks/{webHookId}", appCr.WebHookGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-hooks/{webHookId}", appCr.WebHookUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-hooks/{webHookId}", appCr.WebHookDelete).Methods(http.MethodDelete, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-hooks/{webHookId}/call", appCr.WebHookCall).Methods(http.MethodGet, http.MethodOptions)

	// ### Article ###
	rAuthFull.HandleFunc("/accounts/{accountHashId}/articles", appCr.ArticleCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/articles", appCr.ArticleListPaginationGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/articles/{articleId:[0-9]+}", appCr.ArticleGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/articles/{articleId:[0-9]+}", appCr.ArticleUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/articles/{articleId:[0-9]+}", appCr.ArticleDelete).Methods(http.MethodDelete, http.MethodOptions)

}
