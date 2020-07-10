package routes

import (
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/controllers/uiApiCr"
	"net/http"
)

/**
* [UI-API] - группа роутов для работы публичного UI/API через ui.api.ratuscrm.com
*
* В контексте issuerAccountId = accountId всегда, т.к. доступ к нескольким аккаунтам не предусматриваются.

* Оба роутера монтируются в точку /accounts/{accountId} имеют в контексте account & accountId
* rUiApi - маршруты без проверки JWT
* rUiApiAuth - маршрут с проверкой JWT, а также на совпадение {accountId} с accountId указаном в токене
 */
//var UiApiRoutes = func (rFree, rUiApiAuthFull *mux.Router) {
var UiApiRoutes = func (rFree *mux.Router) {

	// rFree.HandleFunc("/users", appCr.UserRegistration).Methods(http.MethodPost, http.MethodOptions)
	// rFree.HandleFunc("/users/auth/username", appCr.UserAuthByUsername).Methods(http.MethodPost, http.MethodOptions)
	// rFree.HandleFunc("/users/auth/email", appCr.UserAuthByEmail).Methods(http.MethodPost, http.MethodOptions)
	
	// rFree.HandleFunc("/accounts/{accountHashId}/shops/{shopId:[0-9]+}/deliveries", uiApiCr.DeliveryGetListByShop).Methods(http.MethodGet, http.MethodOptions)
	rFree.HandleFunc("/shops/{shopId:[0-9]+}/deliveries", uiApiCr.DeliveryGetListByShop).Methods(http.MethodGet, http.MethodOptions)

	rFree.HandleFunc("/shops/{shopId:[0-9]+}/deliveries-calculate", uiApiCr.DeliveryCalculateDeliveryCost).Methods(http.MethodPost, http.MethodOptions)

	rFree.HandleFunc("/shops/{shopId:[0-9]+}/deliveries-list-options", uiApiCr.DeliveryListOptions).Methods(http.MethodGet, http.MethodOptions)


}