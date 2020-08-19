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
	rAuthFull.HandleFunc("/accounts/{accountHashId}/users/upload", appCr.UserUpload).Methods(http.MethodPost, http.MethodOptions)
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
	// rAuthFull.HandleFunc("/accounts/{accountHashId}/email-templates/{emailTemplateHashId}/send/user", appCr.EmailTemplateSendToUser).Methods(http.MethodPost, http.MethodOptions)

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
	rAuthFull.HandleFunc("/accounts/{accountHashId}/products", appCr.ProductCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/products/{productId:[0-9]+}", appCr.ProductGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/products", appCr.ProductListPaginationGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/products/{productId:[0-9]+}", appCr.ProductUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/products/{productId:[0-9]+}", appCr.ProductDelete).Methods(http.MethodDelete, http.MethodOptions)

	// ### Deliveries ####
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/deliveries", appCr.DeliveryCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/deliveries", appCr.DeliveryGetListByShop).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/deliveries", appCr.DeliveryUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/deliveries", appCr.DeliveryDelete).Methods(http.MethodDelete, http.MethodOptions)

	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-sites/{webSiteId:[0-9]+}/deliveries-code-list", uiApiCr.DeliveryCodeList).Methods(http.MethodGet, http.MethodOptions)

	// ### WebHooks ####
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-hooks", appCr.WebHookCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-hooks", appCr.WebHookListPaginationGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-hooks/{webHookId}", appCr.WebHookGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-hooks/{webHookId}", appCr.WebHookUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-hooks/{webHookId}", appCr.WebHookDelete).Methods(http.MethodDelete, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/web-hooks/{webHookId}/execute", appCr.WebHookExecute).Methods(http.MethodGet, http.MethodOptions)

	// ### EmailBoxes ####
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-boxes", appCr.EmailBoxCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-boxes", appCr.EmailBoxListPaginationGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-boxes/{emailBoxId:[0-9]+}", appCr.EmailBoxGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-boxes/{emailBoxId:[0-9]+}", appCr.EmailBoxUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-boxes/{emailBoxId:[0-9]+}", appCr.EmailBoxDelete).Methods(http.MethodDelete, http.MethodOptions)

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

	// ### Email Queue ####
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-queues", appCr.EmailQueueCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-queues", appCr.EmailQueueGetListPagination).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-queues/{emailQueueId:[0-9]+}", appCr.EmailQueueGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-queues/{emailQueueId:[0-9]+}", appCr.EmailQueueUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-queues/{emailQueueId:[0-9]+}", appCr.EmailQueueDelete).Methods(http.MethodDelete, http.MethodOptions)

	// ### Email Campaign ####
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-campaigns", appCr.EmailCampaignCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-campaigns", appCr.EmailCampaignGetListPagination).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-campaigns/{emailCampaignId:[0-9]+}", appCr.EmailCampaignGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-campaigns/{emailCampaignId:[0-9]+}", appCr.EmailCampaignUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-campaigns/{emailCampaignId:[0-9]+}", appCr.EmailCampaignDelete).Methods(http.MethodDelete, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-campaigns/{emailCampaignId:[0-9]+}/planning", appCr.EmailCampaignPlanning).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-campaigns/{emailCampaignId:[0-9]+}/execute", appCr.EmailCampaignExecute).Methods(http.MethodGet, http.MethodOptions)

	// ### Email Queue ####
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-queues/{emailQueueId:[0-9]+}/email-queue-email-templates", appCr.EmailQueueEmailTemplateCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-queues/{emailQueueId:[0-9]+}/email-queue-email-templates", appCr.EmailQueueEmailTemplateGetListPagination).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-queues/{emailQueueId:[0-9]+}/email-queue-email-templates", appCr.EmailQueueEmailTemplateMassUpdates).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-queues/{emailQueueId:[0-9]+}/email-queue-email-templates/{emailQueueEmailTemplateId:[0-9]+}", appCr.EmailQueueEmailTemplateGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-queues/{emailQueueId:[0-9]+}/email-queue-email-templates/{emailQueueEmailTemplateId:[0-9]+}", appCr.EmailQueueEmailTemplateUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/email-queues/{emailQueueId:[0-9]+}/email-queue-email-templates/{emailQueueEmailTemplateId:[0-9]+}", appCr.EmailQueueEmailTemplateDelete).Methods(http.MethodDelete, http.MethodOptions)

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

	// ### Order Statuses ###
	rAuthFull.HandleFunc("/accounts/{accountHashId}/order-statuses", appCr.OrderStatusCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/order-statuses", appCr.OrderStatusGetList).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/order-statuses/{orderStatusId:[0-9]+}", appCr.OrderStatusGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/order-statuses/{orderStatusId:[0-9]+}", appCr.OrderStatusUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/order-statuses/{orderStatusId:[0-9]+}", appCr.OrderStatusDelete).Methods(http.MethodDelete, http.MethodOptions)

	// ### Payments ###
	rAuthFull.HandleFunc("/accounts/{accountHashId}/payments", appCr.PaymentGetListPagination).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/payments/{paymentId:[0-9]+}", appCr.PaymentGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/payments/{paymentId:[0-9]+}", appCr.PaymentUpdate).Methods(http.MethodPatch, http.MethodOptions)

	// ### Payment Subject Items (SYSTEM) ###
	rAuthFull.HandleFunc("/accounts/{accountHashId}/payment-subjects", appCr.PaymentSubjectCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/payment-subjects", appCr.PaymentSubjectGetListPagination).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/payment-subjects/{paymentSubjectsId:[0-9]+}", appCr.PaymentSubjectGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/payment-subjects/{paymentSubjectsId:[0-9]+}", appCr.PaymentSubjectUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/payment-subjects/{paymentSubjectsId:[0-9]+}", appCr.PaymentSubjectDelete).Methods(http.MethodDelete, http.MethodOptions)

	// ### Payment Mode (SYSTEM) ###
	rAuthFull.HandleFunc("/accounts/{accountHashId}/payment-modes", appCr.PaymentModeCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/payment-modes", appCr.PaymentModeGetList).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/payment-modes/{paymentModeId:[0-9]+}", appCr.PaymentModeGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/payment-modes/{paymentModeId:[0-9]+}", appCr.PaymentModeUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/payment-modes/{paymentModeId:[0-9]+}", appCr.PaymentModeDelete).Methods(http.MethodDelete, http.MethodOptions)

	// ### Payment Method Items ###
	rAuthFull.HandleFunc("/accounts/{accountHashId}/payment-methods", appCr.PaymentMethodCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/payment-methods", appCr.PaymentMethodGetList).Methods(http.MethodGet, http.MethodOptions)
	// требуется указать ?code='cash' / ?code='yandex' /
	rAuthFull.HandleFunc("/accounts/{accountHashId}/payment-methods/{paymentMethodId:[0-9]+}", appCr.PaymentMethodGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/payment-methods/{paymentMethodId:[0-9]+}", appCr.PaymentMethodUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/payment-methods/{paymentMethodId:[0-9]+}", appCr.PaymentMethodDelete).Methods(http.MethodDelete, http.MethodOptions)

	// ### Vat Code ###
	rAuthFull.HandleFunc("/accounts/{accountHashId}/vat-codes", appCr.VatCodeCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/vat-codes", appCr.VatCodeGetListPagination).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/vat-codes/{vatCodeId:[0-9]+}", appCr.VatCodeGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/vat-codes/{vatCodeId:[0-9]+}", appCr.VatCodeUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/vat-codes/{vatCodeId:[0-9]+}", appCr.VatCodeDelete).Methods(http.MethodDelete, http.MethodOptions)

	// ### Delivery Order ###
	rAuthFull.HandleFunc("/accounts/{accountHashId}/delivery-orders", appCr.DeliveryOrderCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/delivery-orders", appCr.DeliveryOrderGetListPagination).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/delivery-orders/{deliveryOrderId:[0-9]+}", appCr.DeliveryOrderGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/delivery-orders/{deliveryOrderId:[0-9]+}", appCr.DeliveryOrderUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/delivery-orders/{deliveryOrderId:[0-9]+}", appCr.DeliveryOrderDelete).Methods(http.MethodDelete, http.MethodOptions)

	// ### Delivery Statuses ###
	rAuthFull.HandleFunc("/accounts/{accountHashId}/delivery-statuses", appCr.DeliveryStatusCreate).Methods(http.MethodPost, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/delivery-statuses", appCr.DeliveryStatusGetList).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/delivery-statuses/{deliveryStatusId:[0-9]+}", appCr.DeliveryStatusGet).Methods(http.MethodGet, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/delivery-statuses/{deliveryStatusId:[0-9]+}", appCr.DeliveryStatusUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rAuthFull.HandleFunc("/accounts/{accountHashId}/delivery-statuses/{deliveryStatusId:[0-9]+}", appCr.DeliveryStatusDelete).Methods(http.MethodDelete, http.MethodOptions)
}
