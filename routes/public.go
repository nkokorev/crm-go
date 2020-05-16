package routes

import (
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/controllers"
	"net/http"
)

// subdomain: public.
var PublicRoutes = func(r *mux.Router) {

	// http://public.crm.local/email/templates/share/4fgjy6lk1kxp
	r.HandleFunc("/email/templates/share/{emailTemplateHashId}", controllers.EmailTemplateShareGet).Methods(http.MethodGet, http.MethodOptions)
}
