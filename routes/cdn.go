package routes

import (
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/controllers"
	"net/http"
)

// subdomain: cdn.<ratuscrm.com>
var CDNRoutes = func(r *mux.Router) {

	// тут надо бы сделать методичку одного окна, чтобы можно было посмотреть и raw и html
	// http://public.crm.local/email/templates/share/4fgjy6lk1kxp
	r.HandleFunc("/emails/preview/raw/{emailTemplateHashId}", controllers.EmailTemplatePreviewGetRawHTML).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/emails/preview/compile/{emailTemplateHashId}", controllers.EmailTemplatePreviewGetHTML).Methods(http.MethodGet, http.MethodOptions)
	// r.HandleFunc("/emails/preview/html/{emailTemplateHashId}", controllers.EmailTemplatePreviewHTMLGet).Methods(http.MethodGet, http.MethodOptions)
	// r.HandleFunc("/emails/preview/raw/{emailTemplateHashId}", controllers.EmailTemplatePreviewRawGet).Methods(http.MethodGet, http.MethodOptions)

	
	// r.HandleFunc("/storage", controllers.StorageGetList).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/storage/{hashId}", controllers.StorageCDNGet).Methods(http.MethodGet, http.MethodOptions)

}
