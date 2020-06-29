package routes

import (
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/controllers"
	"net/http"
)

/**
* [API] - группа роутов доступных только после Bearer Авторизации. В контексте всегда доступен account & accountId
*/
var ApiRoutes = func (rApi *mux.Router) {

	// загружаем базовые настройки системы
	rApi.HandleFunc("/", controllers.CheckApi).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)

	rApi.HandleFunc("/users", controllers.CreateUser).Methods(http.MethodPost)

	rApi.HandleFunc("/shops", controllers.ShopListGet).Methods(http.MethodGet)
	rApi.HandleFunc("/shops/{shopId:[0-9]+}", controllers.ShopGet).Methods(http.MethodGet, http.MethodOptions)

	rApi.HandleFunc("/shops/{shopId:[0-9]+}/product-groups", controllers.ProductGroupListPaginationByShopGet).Methods(http.MethodGet)
	rApi.HandleFunc("/shops/{shopId:[0-9]+}/product-groups/{productGroupId:[0-9]+}", controllers.ProductGroupByShopGet).Methods(http.MethodGet)

	rApi.HandleFunc("/shops/{shopId:[0-9]+}/product-cards", controllers.ProductCardListPaginationByShopGet).Methods(http.MethodGet)
	rApi.HandleFunc("/shops/{shopId:[0-9]+}/product-cards/{productCardId:[0-9]+}", controllers.ProductCardByShopGet).Methods(http.MethodGet)

	rApi.HandleFunc("/shops/{shopId:[0-9]+}/products", controllers.ProductListPaginationGet).Methods(http.MethodGet)
	rApi.HandleFunc("/shops/{shopId:[0-9]+}/products/{productId:[0-9]+}", controllers.ProductGet).Methods(http.MethodGet)

	rApi.HandleFunc("/articles", controllers.ArticleListPaginationGet).Methods(http.MethodGet)
	rApi.HandleFunc("/articles/{articleId:[0-9]+}", controllers.ArticleGet).Methods(http.MethodGet)

		
}