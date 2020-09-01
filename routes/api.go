package routes

import (
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/controllers/appCr"
	"net/http"
)

/**
* [API] - группа роутов доступных только после Bearer Авторизации. В контексте всегда доступен account & accountId

*/
var ApiRoutes = func (rApi *mux.Router) {

	// загружаем базовые настройки системы
	rApi.HandleFunc("/", appCr.CheckApi).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)

	rApi.HandleFunc("/web-sites", appCr.WebSiteListGet).Methods(http.MethodGet)
	rApi.HandleFunc("/web-sites/{webSiteId:[0-9]+}", appCr.WebSiteGet).Methods(http.MethodGet, http.MethodOptions)

	// rApi.HandleFunc("/web-sites/{webSiteId:[0-9]+}/product-groups", appCr.ProductGroupListPaginationByShopGet).Methods(http.MethodGet)
	// rApi.HandleFunc("/web-sites/{webSiteId:[0-9]+}/product-groups/{productGroupId:[0-9]+}", appCr.ProductGroupByShopGet).Methods(http.MethodGet)

	rApi.HandleFunc("/product-cards", appCr.ProductCardListPaginationGet).Methods(http.MethodGet)
	rApi.HandleFunc("/product-cards/{productCardId:[0-9]+}", appCr.ProductCardGet).Methods(http.MethodGet)

	rApi.HandleFunc("/web-sites/{webSiteId:[0-9]+}/products", appCr.ProductListPaginationGet).Methods(http.MethodGet)
	rApi.HandleFunc("/web-sites/{webSiteId:[0-9]+}/products/{productId:[0-9]+}", appCr.ProductGet).Methods(http.MethodGet)

	rApi.HandleFunc("/articles", appCr.ArticleListPaginationGet).Methods(http.MethodGet)
	rApi.HandleFunc("/articles/{articleId:[0-9]+}", appCr.ArticleGet).Methods(http.MethodGet)

	// ######### User #########
	// rApi.HandleFunc("/articles/{articleId:[0-9]+}", appCr.ArticleGet).Methods(http.MethodGet)
	rApi.HandleFunc("/users", appCr.UserCreate).Methods(http.MethodPost, http.MethodOptions)
	rApi.HandleFunc("/users", appCr.UsersGetListPagination).Methods(http.MethodGet, http.MethodOptions)
	rApi.HandleFunc("/users/{userId:[0-9]+}", appCr.UserGet).Methods(http.MethodGet, http.MethodOptions)
	rApi.HandleFunc("/users/{userId:[0-9]+}", appCr.UserUpdate).Methods(http.MethodPatch, http.MethodOptions)
	rApi.HandleFunc("/users/{userId:[0-9]+}", appCr.UserRemoveFromAccount).Methods(http.MethodDelete, http.MethodOptions)

}