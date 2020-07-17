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

	rApi.HandleFunc("/shops", appCr.ShopListGet).Methods(http.MethodGet)
	rApi.HandleFunc("/shops/{shopId:[0-9]+}", appCr.ShopGet).Methods(http.MethodGet, http.MethodOptions)

	rApi.HandleFunc("/shops/{shopId:[0-9]+}/product-groups", appCr.ProductGroupListPaginationByShopGet).Methods(http.MethodGet)
	rApi.HandleFunc("/shops/{shopId:[0-9]+}/product-groups/{productGroupId:[0-9]+}", appCr.ProductGroupByShopGet).Methods(http.MethodGet)

	rApi.HandleFunc("/shops/{shopId:[0-9]+}/product-cards", appCr.ProductCardListPaginationByShopGet).Methods(http.MethodGet)
	rApi.HandleFunc("/shops/{shopId:[0-9]+}/product-cards/{productCardId:[0-9]+}", appCr.ProductCardByShopGet).Methods(http.MethodGet)

	rApi.HandleFunc("/shops/{shopId:[0-9]+}/products", appCr.ProductListPaginationGet).Methods(http.MethodGet)
	rApi.HandleFunc("/shops/{shopId:[0-9]+}/products/{productId:[0-9]+}", appCr.ProductGet).Methods(http.MethodGet)

	rApi.HandleFunc("/articles", appCr.ArticleListPaginationGet).Methods(http.MethodGet)
	rApi.HandleFunc("/articles/{articleId:[0-9]+}", appCr.ArticleGet).Methods(http.MethodGet)

	// ######### User #########
	// rApi.HandleFunc("/articles/{articleId:[0-9]+}", appCr.ArticleGet).Methods(http.MethodGet)


}