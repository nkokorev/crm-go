package routes

import (
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/controllers/appCr"
	"net/http"
)

/**
* [TRACKING] - группа роутов для отслеживания внешних данных
* MiddleWare - Cors & AccountHashId
*/
var TrackingRoutes = func (r *mux.Router) {

	// загружаем базовые настройки системы
	r.HandleFunc("/", appCr.CheckApi).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)
	r.HandleFunc("/", appCr.CheckApi).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)
}