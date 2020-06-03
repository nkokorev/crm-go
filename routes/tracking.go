package routes

import (
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/controllers"
	"net/http"
)

/**
* [TRACKING] - группа роутов для отслеживания внешних данных
*/
var TrackingRoutes = func (r *mux.Router) {

	// загружаем базовые настройки системы
	r.HandleFunc("/", controllers.CheckApi).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)
}