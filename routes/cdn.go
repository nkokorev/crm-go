package routes

import (
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/controllers/appCr"
	"net/http"
)

// subdomain: cdn.<ratuscrm.com>
var CDNRoutes = func(r *mux.Router) {

	// тут надо бы сделать методичку одного окна, чтобы можно было посмотреть и raw и html
	// http://public.crm.local/email/templates/share/4fgjy6lk1kxp
	r.HandleFunc("/emails/preview/raw/{emailTemplateHashId}", appCr.EmailTemplatePreviewGetRawHTML).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/emails/preview/compile/{emailTemplateHashId}", appCr.EmailTemplatePreviewGetHTML).Methods(http.MethodGet, http.MethodOptions)
	// r.HandleFunc("/emails/preview/html/{emailTemplateHashId}", controllers.EmailTemplatePreviewHTMLGet).Methods(http.MethodGet, http.MethodOptions)
	// r.HandleFunc("/emails/preview/raw/{emailTemplateHashId}", controllers.EmailTemplatePreviewRawGet).Methods(http.MethodGet, http.MethodOptions)

	// r.HandleFunc("/storage", controllers.StorageGetList).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/public/{hashId}", appCr.StorageCDNGet).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/products/images/{hashId}", appCr.StorageCDNGet).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/emails/images/{hashId}", appCr.StorageCDNGet).Methods(http.MethodGet, http.MethodOptions)

	r.HandleFunc("/articles/preview/raw/{hashId}", appCr.ArticleRawPreviewCDNGet).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/articles/preview/compile/{hashId}", appCr.ArticleCompilePreviewCDNGet).Methods(http.MethodGet, http.MethodOptions)

}
