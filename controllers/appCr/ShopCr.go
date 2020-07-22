package appCr

import (
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)



/////////////////////////////////////

func ProductGroupCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	webSiteID, err := utilsCr.GetUINTVarFromRequest(r, "webSiteID")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}

	var webSite models.WebSite
	err = account.LoadEntity(&webSite, webSiteID)
	// webSite, err := account.GetShop(webSiteID)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	// Get JSON-request
	var input struct{
		models.ProductGroup
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	group, err := webSite.CreateProductGroup(input.ProductGroup)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания группы"}))
		return
	}

	resp := u.Message(true, "POST ProductGroup Created")
	resp["group"] = *group
	u.Respond(w, resp)
}

func ProductGroupByShopGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	webSiteID, err := utilsCr.GetUINTVarFromRequest(r, "webSiteID")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}

	productGroupID, err := utilsCr.GetUINTVarFromRequest(r, "productGroupID")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID product group"))
		return
	}

	var webSite models.WebSite
	err = account.LoadEntity(&webSite, webSiteID)
	// webSite, err := account.GetShop(webSiteID)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	productGroup, err := webSite.GetProductGroup(productGroupID)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
		return
	}

	resp := u.Message(true, "GET Product Group")
	resp["productGroup"] = productGroup
	u.Respond(w, resp)
}
func ProductGroupListPaginationByShopGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	webSiteID, err := utilsCr.GetUINTVarFromRequest(r, "webSiteID")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}

	var webSite models.WebSite
	err = account.LoadEntity(&webSite, webSiteID)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	limit, ok := utilsCr.GetQueryINTVarFromGET(r, "limit")
	if !ok {
		limit = 25
	}
	offset, ok := utilsCr.GetQueryINTVarFromGET(r, "offset")
	if !ok || offset < 0 {
		offset = 0
	}
	sortDesc := utilsCr.GetQueryBoolVarFromGET(r, "sortDesc") // обратный или нет порядок
	sortBy, ok := utilsCr.GetQuerySTRVarFromGET(r, "sortBy")
	if !ok {
		sortBy = ""
	}
	if sortDesc {
		sortBy += " desc"
	}
	search, ok := utilsCr.GetQuerySTRVarFromGET(r, "search")
	if !ok {
		search = ""
	}
	// 2. Узнаем, какой список нужен
	all, allOk := utilsCr.GetQuerySTRVarFromGET(r, "all")


	var total uint = 0
	productGroups := make([]models.ProductGroup,0)

	if all == "true" && allOk {

		// webSites, total, err = account.GetListEntity(&models.ProductGroup{}, sortBy)
		productGroups, err = webSite.GetProductGroupList()
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список сайтов"))
			return
		}
	} else {
		// webHooks, total, err = account.GetWebHooksPaginationList(offset, limit, search)
		// webSites, total, err = account.GetPaginationListEntity(&models.ProductGroup{}, offset, limit, sortBy, search)
		productGroups, total, err = webSite.GetProductGroupsPaginationList(offset, limit, search)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список сайтов"))
			return
		}
	}


	resp := u.Message(true, "GET ProductGroup PaginationList")
	resp["productGroups"] = productGroups
	resp["total"] = total
	u.Respond(w, resp)
}

func ProductGroupListGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	productGroups, err := account.GetProductGroups()
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
		return
	}

	resp := u.Message(true, "GET Product Group List")
	resp["productGroups"] = productGroups
	u.Respond(w, resp)
}

func ProductGroupUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	webSiteID, err := utilsCr.GetUINTVarFromRequest(r, "webSiteID")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}

	var webSite models.WebSite
	err = account.LoadEntity(&webSite, webSiteID)
	// webSite, err := account.GetShop(webSiteID)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	productGroupID, err := utilsCr.GetUINTVarFromRequest(r, "productGroupID")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID группы"))
		return
	}

	// var input interface{}
	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// group, err := webSite.UpdateProductGroup(groupID, &input.ProductGroup)
	group, err := webSite.UpdateProductGroup(productGroupID, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH ProductGroup Update")
	resp["group"] = group
	u.Respond(w, resp)
}

func ProductGroupDelete(w http.ResponseWriter, r *http.Request) {
	
	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	webSiteID, err := utilsCr.GetUINTVarFromRequest(r, "webSiteID")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}

	var webSite models.WebSite
	err = account.LoadEntity(&webSite, webSiteID)
	// webSite, err := account.GetShop(webSiteID)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	productGroupID, err := utilsCr.GetUINTVarFromRequest(r, "productGroupID")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}


	if err = webSite.DeleteProductGroup(productGroupID); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении товарной группы"))
		return
	}

	resp := u.Message(true, "DELETE ProductGroup Successful")
	u.Respond(w, resp)
}


/////////////////////////////////////

func ProductCardByShopCreate(w http.ResponseWriter, r *http.Request) {
	
	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	webSiteID, err := utilsCr.GetUINTVarFromRequest(r, "webSiteID")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}

	var webSite models.WebSite
	err = account.LoadEntity(&webSite, webSiteID)
	// webSite, err := account.GetShop(webSiteID)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	// Get JSON-request
	var input struct{
		models.ProductCard
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}
	
	// special fix!
	input.ProductCard.WebSiteID = webSiteID

	card, err := webSite.CreateProductCard(input.ProductCard, nil)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания группы"}))
		return
	}

	resp := u.Message(true, "POST ProductCard Created")
	resp["card"] = *card
	u.Respond(w, resp)
}

func ProductCardCreate(w http.ResponseWriter, r *http.Request) {


	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	var input struct{
		models.ProductCard
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	card, err := account.CreateProductCard(input.ProductCard)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания группы"}))
		return
	}

	resp := u.Message(true, "POST ProductCard Created")
	resp["card"] = *card
	u.Respond(w, resp)
}

// Собираем все картоки товаров для конкретного магазина
func ProductCardByShopGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	productCardID, err := utilsCr.GetUINTVarFromRequest(r, "productCardID")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке product Card ID"))
		return
	}

	webSiteID, err := utilsCr.GetUINTVarFromRequest(r, "webSiteID")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}

	var webSite models.WebSite
	err = account.LoadEntity(&webSite, webSiteID)
	// webSite, err := account.GetShop(webSiteID)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	productCard, err := webSite.GetProductCard(productCardID)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить карточку товара"))
		return
	}

	resp := u.Message(true, "GET Product Card")
	resp["productCard"] = productCard
	u.Respond(w, resp)
}

func ProductCardListPaginationByShopGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	webSiteID, err := utilsCr.GetUINTVarFromRequest(r, "webSiteID")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}

	var webSite models.WebSite
	err = account.LoadEntity(&webSite, webSiteID)
	// webSite, err := account.GetShop(webSiteID)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	// 2. Узнаем, какой список нужен
	limit, ok := utilsCr.GetQueryINTVarFromGET(r, "limit")
	if !ok {
		limit = 100
	}
	offset, ok := utilsCr.GetQueryINTVarFromGET(r, "offset")
	if !ok || offset < 0 {
		offset = 0
	}
	search, ok := utilsCr.GetQuerySTRVarFromGET(r, "search")
	if !ok {
		search = ""
	}
	products, _ := utilsCr.GetQuerySTRVarFromGET(r, "products")

	productCards, total, err := webSite.GetProductCardList(offset, limit, search, products == "true")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
		return
	}

	resp := u.Message(true, "GET Product Card List")
	resp["total"] = total
	resp["productCards"] = productCards
	u.Respond(w, resp)
}

func ProductCardUpdate(w http.ResponseWriter, r *http.Request) {

	fmt.Println("ss")
	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	productCardID, err := utilsCr.GetUINTVarFromRequest(r, "productCardID")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID группы"))
		return
	}

	// var input interface{}
	var input map[string]interface{}
	/*var input = struct {
		models.ProductCard
	}{}*/

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	if input["switchProducts"] == nil {
		input["switchProducts"] = pq.StringArray{"color"}
	}


	card, err := account.UpdateProductCard(productCardID, input)
	//card, err := account.UpdateProductCard(cardID, input.ProductCard)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Product Card Update")
	resp["card"] = card
	u.Respond(w, resp)
}

func ProductCardDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productCardID, err := utilsCr.GetUINTVarFromRequest(r, "productCardID")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}


	if err = account.DeleteProductCard(productCardID); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении карточки товара"))
		return
	}

	resp := u.Message(true, "DELETE Product Card Successful")
	u.Respond(w, resp)
}

/////////////////////////////////////

func ProductCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	var input struct{
		models.Product
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	product, err := account.CreateProduct(input.Product)
	if err != nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Ошибка во время создания продукта"))
		return
	}

	resp := u.Message(true, "POST Product Created")
	resp["product"] = *product
	u.Respond(w, resp)
}

func ProductGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	productID, err := utilsCr.GetUINTVarFromRequest(r, "productID")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID товара"))
		return
	}

	product, err := account.GetProduct(productID)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
		return
	}

	resp := u.Message(true, "GET Product")
	resp["product"] = product
	u.Respond(w, resp)
}


func ProductListPaginationGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	// 2. Узнаем, какой список нужен
	limit, ok := utilsCr.GetQueryINTVarFromGET(r, "limit")
	if !ok {
		limit = 100
	}
	offset, ok := utilsCr.GetQueryINTVarFromGET(r, "offset")
	if !ok || offset < 0 {
		offset = 0
	}
	search, ok := utilsCr.GetQuerySTRVarFromGET(r, "search")
	if !ok {
		search = ""
	}

	products, total, err := account.GetProductListPagination(offset, limit, search)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
		return
	}

	resp := u.Message(true, "GET Product List Pagination")
	resp["total"] = total
	resp["products"] = products
	u.Respond(w, resp)
}

func ProductUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	productID, err := utilsCr.GetUINTVarFromRequest(r, "productID")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID товара"))
		return
	}

	var input map[string]interface{}
	/*var input = struct {
		models.Product	`json:"product"`
		Attributes map[string]interface{} `json:"attributes"`
	}{}*/

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	//card, err := account.UpdateProduct(productID, input.Product)
	card, err := account.UpdateProduct(productID, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Product Update")
	resp["product"] = card
	u.Respond(w, resp)
}

func ProductDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productID, err := utilsCr.GetUINTVarFromRequest(r, "productID")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID товара"))
		return
	}


	if err = account.DeleteProduct(productID); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении карточки товара"))
		return
	}

	resp := u.Message(true, "DELETE Product Successful")
	u.Respond(w, resp)
}

////////////////////////////////////

func ProductAttributeList(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	/*attrs, err := account.GetProductListPagination(offset, limit, search)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
		return
	}*/

	resp := u.Message(true, "GET Product List Pagination")
	// resp["total"] = total
	// resp["products"] = products
	u.Respond(w, resp)
}