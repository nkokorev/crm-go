package routes

import (
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/controllers/uiApiCr"

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
	// rAuthFull.HandleFunc("/accounts/{accountHashId}/users", appCr.UsersGetListPagination).Methods(http.MethodGet, http.MethodOptions)
	// rAuthFull.HandleFunc("/accounts/{accountHashId}/users/{userHashId}", appCr.UserRemoveFromAccount).Methods(http.MethodDelete, http.MethodOptions)
	// rAuthFull.HandleFunc("/accounts/{accountHashId}/users/{userHashId}", appCr.UserUpdate).Methods(http.MethodPatch, http.MethodOptions)

	rAuthFull.HandleFunc("/accounts/{accountHashId}/users", appCr.UserCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/users", appCr.UsersGetListPagination).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/users/{userId:[0-9]+}", appCr.UserGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/users/{userId:[0-9]+}", appCr.UserUpdate).Methods(http.MethodPatch, http.MethodOptions)

	rAuthFull.HandleFunc("/accounts/{accountHashId}/users/{userId:[0-9]+}", appCr.UserRemoveFromAccount).Methods(http.MethodDelete, http.MethodOptions)

	// rAuthFull.HandleFunc("/accounts/{accountHashId}/users/{userId:[0-9]+}", appCr.UserDelete).Methods(http.MethodDelete, http.MethodOptions)


	// -- ROLES --
	// Запрос ниже может иметь много параметров (диапазон выборки, число пользователей)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/roles", appCr.RoleGetList).Methods(http.MethodGet, http.MethodOptions)

	rAuthUser.HandleFunc("/accounts/{accountHashId}/auth", appCr.AccountAuthUser).Methods(http.MethodPost, http.MethodOptions)

	// ######## Uses #########
	// -- CRUD --
	rAuthUser.HandleFunc("/users/{hashId}/app/auth/username", appCr.UserAccountsGet).Methods(http.MethodGet, http.MethodOptions)
	//rAuthUser.HandleFunc("/users/accounts", appCr.UserAccountsGet).Methods(http.MethodGet, http.MethodOptions)

	// ### ApiKeys ###
	// -- CRUD --
	rAuthFull.HandleFunc("/accounts/{accountHashId}/api-keys", appCr.ApiKeyCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/api-keys", appCr.ApiKeyGetList).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/api-keys/{apiKeyId}", appCr.ApiKeyGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/api-keys/{id}", appCr.ApiKeyUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/api-keys/{id}", appCr.ApiKeyDelete).Methods(http.MethodDelete, http.MethodOptions)

	// ---CRUD---

	// ### Email templates ###
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-templates", appCr.EmailTemplateCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-templates", appCr.EmailTemplateGetListPagination).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-templates/{emailTemplateId}", appCr.EmailTemplateGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-templates/{emailTemplateId}", appCr.EmailTemplateUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-templates/{emailTemplateId}", appCr.EmailTemplateDelete).Methods(http.MethodDelete, http.MethodOptions)
	// !!!!!!!!
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-templates/{emailTemplateHashId}/send/user", appCr.EmailTemplateSendToUser).Methods(http.MethodPost, http.MethodOptions)

	// ### STORAGE CRUD ####
	rAuthFull.HandleFunc("/accounts/{accountHashId}/storage", appCr.StorageCreateFile).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/storage", appCr.StorageGetListPagination).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/storage/{fileId}", appCr.StorageGetFile).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/storage/{fileId}", appCr.StorageUpdateFile).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/storage/{fileId}", appCr.StorageDeleteFile).Methods(http.MethodDelete, http.MethodOptions)

	// ### Web Site ####
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites", appCr.WebSiteCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites", appCr.WebSiteListPaginationGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}", appCr.WebSiteGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}", appCr.WebSiteUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}", appCr.WebSiteDelete).Methods(http.MethodDelete, http.MethodOptions)

	// ### Product Group ####
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/product-groups", appCr.ProductGroupCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/product-groups", appCr.ProductGroupListPaginationByShopGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/product-groups/{productGroupId:[0-9]+}", appCr.ProductGroupByShopGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/product-groups/{productGroupId:[0-9]+}", appCr.ProductGroupUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/product-groups/{productGroupId:[0-9]+}", appCr.ProductGroupDelete).Methods(http.MethodDelete, http.MethodOptions)

	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/product-cards", appCr.ProductCardCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/product-cards", appCr.ProductCardByShopCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/product-cards/{productCardId:[0-9]+}", appCr.ProductCardByShopGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/product-cards", appCr.ProductCardListPaginationByShopGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/product-cards/{productCardId:[0-9]+}", appCr.ProductCardUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/product-cards/{productCardId:[0-9]+}", appCr.ProductCardDelete).Methods(http.MethodDelete, http.MethodOptions)

	// todo: byShop - сейчас косвенная принадлежность товаров к магазину не отслеживается
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/products", appCr.ProductCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/products/{productId:[0-9]+}", appCr.ProductGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/products", appCr.ProductListPaginationGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/products/{productId:[0-9]+}", appCr.ProductUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/products/{productId:[0-9]+}", appCr.ProductDelete).Methods(http.MethodDelete, http.MethodOptions)

	// ### Deliveries ####
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/deliveries", appCr.DeliveryCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/deliveries", appCr.DeliveryGetListByShop).Methods(http.MethodGet, http.MethodOptions)
	// rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/deliveries/russianPost", appCr.DeliveryUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/deliveries", appCr.DeliveryUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/deliveries", appCr.DeliveryDelete).Methods(http.MethodDelete, http.MethodOptions)

	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/deliveries-list-options", uiApiCr.DeliveryListOptions).Methods(http.MethodGet, http.MethodOptions)

	// ### WebHooks ####
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-hooks", appCr.WebHookCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-hooks", appCr.WebHookListPaginationGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-hooks/{webHookId}", appCr.WebHookGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-hooks/{webHookId}", appCr.WebHookUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-hooks/{webHookId}", appCr.WebHookDelete).Methods(http.MethodDelete, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-hooks/{webHookId}/execute", appCr.WebHookExecute).Methods(http.MethodGet, http.MethodOptions)

	// ### EmailBoxes ####
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/email-boxes", appCr.EmailBoxCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/email-boxes", appCr.EmailBoxListGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/email-boxes", appCr.EmailBoxFullListGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/email-boxes/{emailBoxId:[0-9]+}", appCr.EmailBoxGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/email-boxes/{emailBoxId:[0-9]+}", appCr.EmailBoxUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/email-boxes/{emailBoxId:[0-9]+}", appCr.EmailBoxDelete).Methods(http.MethodDelete, http.MethodOptions)

	// ### Article ###
	rAuthFull.HandleFunc("/accounts/{accountHashId}/articles", appCr.ArticleCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/articles", appCr.ArticleListPaginationGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/articles/{articleId:[0-9]+}", appCr.ArticleGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/articles/{articleId:[0-9]+}", appCr.ArticleUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/articles/{articleId:[0-9]+}", appCr.ArticleDelete).Methods(http.MethodDelete, http.MethodOptions)// ### Article ###

	// ### EventHandlers ###
	rAuthFull.HandleFunc("/accounts/{accountHashId}/event-listeners", appCr.EventListenerCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/event-listeners", appCr.EventListenerGetListPagination).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/event-listeners/{eventListenerId:[0-9]+}", appCr.EventListenerGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/event-listeners/{eventListenerId:[0-9]+}", appCr.EventListenerUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/event-listeners/{eventListenerId:[0-9]+}", appCr.EventListenerDelete).Methods(http.MethodDelete, http.MethodOptions)

	// rAuthFull.HandleFunc("/accounts/{accountHashId}/events", appCr.EventSystemListGet).Methods(http.MethodGet, http.MethodOptions)
	// rAuthFull.HandleFunc("/accounts/{accountHashId}/handlers", appCr.HandlersSystemListGet).Methods(http.MethodGet, http.MethodOptions)

	// ### Handler Items ###
	rAuthFull.HandleFunc("/accounts/{accountHashId}/handlers", appCr.HandlerItemCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/handlers", appCr.HandlerItemGetListPagination).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/handlers/{handlerItemId:[0-9]+}", appCr.HandlerItemGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/handlers/{handlerItemId:[0-9]+}", appCr.HandlerItemUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/handlers/{handlerItemId:[0-9]+}", appCr.HandlerItemDelete).Methods(http.MethodDelete, http.MethodOptions)

	// ### Event items ###
	rAuthFull.HandleFunc("/accounts/{accountHashId}/events", appCr.EventItemCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/events", appCr.EventItemGetListPagination).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/events/{eventItemId:[0-9]+}", appCr.EventItemGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/events/{eventItemId:[0-9]+}", appCr.EventItemUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/events/{eventItemId:[0-9]+}", appCr.EventItemDelete).Methods(http.MethodDelete, http.MethodOptions)

	// ### Email Templates ####
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-notifications", appCr.EmailNotificationCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-notifications", appCr.EmailNotificationGetListPagination).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-notifications/{emailNotificationId}", appCr.EmailNotificationGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-notifications/{emailNotificationId}", appCr.EmailNotificationUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-notifications/{emailNotificationId}", appCr.EmailNotificationDelete).Methods(http.MethodDelete, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-notifications/{emailNotificationId}/execute", appCr.EmailNotificationExecute).Methods(http.MethodGet, http.MethodOptions)

	// ### Order Items ###
	rAuthFull.HandleFunc("/accounts/{accountHashId}/orders", appCr.OrderCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/orders", appCr.OrderGetListPagination).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/orders/{orderId:[0-9]+}", appCr.OrderGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/orders/{orderId:[0-9]+}", appCr.OrderUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/orders/{orderId:[0-9]+}", appCr.OrderDelete).Methods(http.MethodDelete, http.MethodOptions)

	// ### Order Channel Items ###
	rAuthFull.HandleFunc("/accounts/{accountHashId}/order-channels", appCr.OrderChannelCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/order-channels", appCr.OrderChannelGetListPagination).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/order-channels/{orderChannelId:[0-9]+}", appCr.OrderChannelGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/order-channels/{orderChannelId:[0-9]+}", appCr.OrderChannelUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/order-channels/{orderChannelId:[0-9]+}", appCr.OrderChannelDelete).Methods(http.MethodDelete, http.MethodOptions)

	// ### Payment Subject Items ###
	rAuthFull.HandleFunc("/accounts/{accountHashId}/payment-subjects", appCr.PaymentSubjectCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/payment-subjects", appCr.PaymentSubjectGetListPagination).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/payment-subjects/{paymentSubjectsId:[0-9]+}", appCr.PaymentSubjectGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/payment-subjects/{paymentSubjectsId:[0-9]+}", appCr.PaymentSubjectUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/payment-subjects/{paymentSubjectsId:[0-9]+}", appCr.PaymentSubjectDelete).Methods(http.MethodDelete, http.MethodOptions)
}
