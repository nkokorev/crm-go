package routes

import (
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/controllers/appCr"
	"net/http"
)

/**
* [TRACKING] - группа роутов для отслеживания внешних данных
*/
var MTA_1_Routes = func (r *mux.Router) {

	// загружаем базовые настройки системы
	r.HandleFunc("/", appCr.CheckApi).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)
}