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
* В контексте issuerAccountId = accountId всегда, т.к. доступ к нескольким аккаунтам не предусматриваются.

* Оба роутера монтируются в точку /accounts/{accountId} имеют в контексте account & accountId
* rUiApi - маршруты без проверки JWT
* rUiApiAuth - маршрут с проверкой JWT, а также на совпадение {accountId} с accountId указаном в токене
 */

// ... /accountHashId}/...
var UiApiRoutesV1 = func (rFree *mux.Router) {

	// rFree.HandleFunc("/users", appCr.UserRegistration).Methods(http.MethodPost, http.MethodOptions)
	// rFree.HandleFunc("/users/auth/username", appCr.UserAuthByUsername).Methods(http.MethodPost, http.MethodOptions)
	// rFree.HandleFunc("/users/auth/email", appCr.UserAuthByEmail).Methods(http.MethodPost, http.MethodOptions)
	
	rFree.HandleFunc("/web-sites/{webSiteId:[0-9]+}/deliveries", uiApiCr.DeliveryGetListByShop).Methods(http.MethodGet, http.MethodOptions)

	rFree.HandleFunc("/web-sites/{webSiteId:[0-9]+}/deliveries-calculate", uiApiCr.DeliveryCalculateDeliveryCost).Methods(http.MethodPost, http.MethodOptions)

	rFree.HandleFunc("/web-sites/{webSiteId:[0-9]+}/deliveries-code-list", uiApiCr.DeliveryCodeList).Methods(http.MethodGet, http.MethodOptions)

	// YandexPayment
	// Адрес для вебхуков от Яндекс.Кассы. Код ответа 200 в случае обработки.
	// rFree.HandleFunc("/payments/yandex-payment/{yandexPayment:[0-9]+}/notifications/", uiApiCr.DeliveryListOptions).Methods(http.MethodGet, http.MethodOptions)

	// URL для яндекса: вставляется hashId магазина, а не id - чтобы защититься от атак.
	rFree.HandleFunc("/yandex-payment/{paymentYandexHashId}/web-hooks", uiApiCr.PaymentYandexWebHook).Methods(http.MethodPost, http.MethodOptions)

	rFree.HandleFunc("/orders", uiApiCr.UiApiOrderCreate).Methods(http.MethodPost, http.MethodOptions)

	// rFree.HandleFunc("/subscribe", uiApiCr.UiApiSubscribe).Methods(http.MethodPost, http.MethodOptions)
	// rFree.HandleFunc("/subscribe", uiApiCr.UiApiSubscribe).Methods(http.MethodPost, http.MethodOptions)
	
	// subscribe & question
	rFree.HandleFunc("/form", uiApiCr.UiApiForm).Methods(http.MethodPost, http.MethodOptions)


	// rFree.HandleFunc("/test", uiApiCr.Test).Methods(http.MethodGet, http.MethodOptions)
	// rFree.NotFoundHandler = NotFoundHandler()
	// rFree.MethodNotAllowedHandler = NotFoundHandler()

	
}
