package routes

import (
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/controllers/appCr"
	"github.com/nkokorev/crm-go/controllers/trackingCr"
	"net/http"
)

/**
* [TRACKING] - группа роутов для отслеживания внешних данных
* MiddleWare - Cors & AccountHashId
*/
var TrackingRoutes = func (rFree *mux.Router) {

	// загружаем базовые настройки системы
	rFree.HandleFunc("/", appCr.CheckApi).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)

	// ?u={userHashId} а аккаунт через accountHashId передан.
	rFree.HandleFunc("/e/unsubscribe", trackingCr.UnsubscribeUser).Methods(http.MethodGet, http.MethodOptions)
	rFree.HandleFunc("/e/open", trackingCr.OpenEmailByPixelUser).Methods(http.MethodGet, http.MethodOptions)

	// pixel ?u={userHashId} а аккаунт через accountHashId передан.
	// r.HandleFunc("/e/open", trackingCr.UnsubscribeUser).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)
}