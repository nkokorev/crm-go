package routes

import (
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/controllers/uiApiCr"
	"net/http"
)

/**
* [UI-API] - группа роутов для работы публичного UI/API через ui.api.ratuscrm.com.
* Авторизации по-умолчанию не требуется!!!
*
* В контексте issuerAccountID = accountID всегда, т.к. доступ к нескольким аккаунтам не предусматриваются.

* Оба роутера монтируются в точку /accounts/{accountID} имеют в контексте account & accountID
* rUiApi - маршруты без проверки JWT
* rUiApiAuth - маршрут с проверкой JWT, а также на совпадение {accountID} с accountID указаном в токене
 */
//var UiApiRoutes = func (rFree, rUiApiAuthFull *mux.Router) {
var UiApiRoutes = func (rFree *mux.Router) {

	// rFree.HandleFunc("/users", appCr.UserRegistration).Methods(http.MethodPost, http.MethodOptions)
	// rFree.HandleFunc("/users/auth/username", appCr.UserAuthByUsername).Methods(http.MethodPost, http.MethodOptions)
	// rFree.HandleFunc("/users/auth/email", appCr.UserAuthByEmail).Methods(http.MethodPost, http.MethodOptions)
	
	rFree.HandleFunc("/web-sites/{webSiteID:[0-9]+}/deliveries", uiApiCr.DeliveryGetListByShop).Methods(http.MethodGet, http.MethodOptions)

	rFree.HandleFunc("/web-sites/{webSiteID:[0-9]+}/deliveries-calculate", uiApiCr.DeliveryCalculateDeliveryCost).Methods(http.MethodPost, http.MethodOptions)

	rFree.HandleFunc("/web-sites/{webSiteID:[0-9]+}/deliveries-list-options", uiApiCr.DeliveryListOptions).Methods(http.MethodGet, http.MethodOptions)


	// YandexPayment
	// Адрес для вебхуков от Яндекс.Кассы. Код ответа 200 в случае обработки.
	// rFree.HandleFunc("/payments/yandex-payment/{yandexPayment:[0-9]+}/notifications/", uiApiCr.DeliveryListOptions).Methods(http.MethodGet, http.MethodOptions)

	// вставляется hashID магазина, а не id - чтобы защититься от атак.
	rFree.HandleFunc("/yandex-payment/{yandexPaymentHashID:[0-9]+}/notifications/", uiApiCr.DeliveryListOptions).Methods(http.MethodGet, http.MethodOptions)

}