package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func ShopCreate(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	var input struct{
		models.Shop
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	shop, err := account.CreateShop(input.Shop)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания ключа"}))
		return
	}

	resp := u.Message(true, "POST Shop Created")
	resp["shop"] = *shop
	u.Respond(w, resp)
}

func ShopListGet(w http.ResponseWriter, r *http.Request) {
	// 1. Получаем рабочий аккаунт (автома. сверка с {hashId}.)
	account, err := GetWorkAccountCheckHashId(w,r)
	if err != nil || account == nil {
		return
	}
	
	shops, err := account.GetShops()
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
		return
	}



	resp := u.Message(true, "GET Shop List")
	resp["shops"] = shops
	u.Respond(w, resp)
}

func ShopUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	shopId, err := GetUINTVarFromRequest(r, "shopId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID шаблона"))
		return
	}

	// Get JSON-request
	var input struct{
		models.Shop
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	shop, err := account.UpdateShop(shopId, &input.Shop)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Shop Update")
	resp["shop"] = shop
	u.Respond(w, resp)
}

func ShopDelete(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	shopId, err := GetUINTVarFromRequest(r, "shopId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID шаблона"))
		return
	}

	if err = account.DeleteShop(shopId); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении магазина"))
		return
	}

	resp := u.Message(true, "DELETE Shop Successful")
	u.Respond(w, resp)
}

/////////////////////////////////////

func ProductGroupCreate(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	shopId, err := GetUINTVarFromRequest(r, "shopId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}

	shop, err := account.GetShop(shopId)
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

	group, err := shop.CreateProductGroup(input.ProductGroup)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания группы"}))
		return
	}

	resp := u.Message(true, "POST ProductGroup Created")
	resp["group"] = *group
	u.Respond(w, resp)
}

func ProductGroupListPaginationByShopGet(w http.ResponseWriter, r *http.Request) {

	var account *models.Account
	var err error
	// 1. Получаем рабочий аккаунт в зависимости от источника (автома. сверка с {hashId}.)
	if isApiRequest(r) {
		account, err = GetWorkAccount(w,r)
		if err != nil || account == nil {
			return
		}
	} else {
		account, err = GetWorkAccountCheckHashId(w,r)
		if err != nil || account == nil {
			return
		}
	}

	shopId, err := GetUINTVarFromRequest(r, "shopId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}

	shop, err := account.GetShop(shopId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	// 2. Узнаем, какой список нужен
	all, allOk := GetQuerySTRVarFromGET(r, "all")

	limit, ok := GetQueryINTVarFromGET(r, "limit")
	if !ok {
		limit = 100
	}
	offset, ok := GetQueryINTVarFromGET(r, "offset")
	if !ok || offset < 0 {
		offset = 0
	}
	search, ok := GetQuerySTRVarFromGET(r, "search")
	if !ok {
		search = ""
	}

	productGroups := make([]models.ProductGroup,0)
	total := 0

	if all == "true" && allOk {
		productGroups, err = shop.GetProductGroups()
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
			return
		}
	} else {
		productGroups, total, err = shop.GetProductGroupsPaginationList(offset, limit, search)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
			return
		}
	}


	resp := u.Message(true, "GET Product Group List")
	resp["productGroups"] = productGroups
	resp["total"] = total
	u.Respond(w, resp)
}

func ProductGroupListGet(w http.ResponseWriter, r *http.Request) {

	// 1. Получаем рабочий аккаунт (автома. сверка с {hashId}.)
	account, err := GetWorkAccountCheckHashId(w,r)
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

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	shopId, err := GetUINTVarFromRequest(r, "shopId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}

	shop, err := account.GetShop(shopId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	groupId, err := GetUINTVarFromRequest(r, "groupId")
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

	// group, err := shop.UpdateProductGroup(groupId, &input.ProductGroup)
	group, err := shop.UpdateProductGroup(groupId, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH ProductGroup Update")
	resp["group"] = group
	u.Respond(w, resp)
}

func ProductGroupDelete(w http.ResponseWriter, r *http.Request) {
	
	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	shopId, err := GetUINTVarFromRequest(r, "shopId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}

	shop, err := account.GetShop(shopId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	groupId, err := GetUINTVarFromRequest(r, "groupId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}


	if err = shop.DeleteProductGroup(groupId); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении товарной группы"))
		return
	}

	resp := u.Message(true, "DELETE ProductGroup Successful")
	u.Respond(w, resp)
}


/////////////////////////////////////

func ProductCardByShopCreate(w http.ResponseWriter, r *http.Request) {
	
	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	shopId, err := GetUINTVarFromRequest(r, "shopId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}

	shop, err := account.GetShop(shopId)
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
	input.ProductCard.ShopID = shopId

	card, err := shop.CreateProductCard(input.ProductCard, nil)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания группы"}))
		return
	}

	resp := u.Message(true, "POST ProductCard Created")
	resp["card"] = *card
	u.Respond(w, resp)
}

func ProductCardCreate(w http.ResponseWriter, r *http.Request) {


	account, err := GetWorkAccount(w,r)
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
func ProductCardListPaginationByShopGet(w http.ResponseWriter, r *http.Request) {

	/*account, err := GetWorkAccountCheckHashId(w,r)
	if err != nil || account == nil {
		return
	}*/
	var account *models.Account
	var err error

	if isApiRequest(r) {
		account, err = GetWorkAccount(w,r)
		if err != nil || account == nil {
			return
		}
	} else {
		account, err = GetWorkAccountCheckHashId(w,r)
		if err != nil || account == nil {
			return
		}
	}

	shopId, err := GetUINTVarFromRequest(r, "shopId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}

	shop, err := account.GetShop(shopId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	// 2. Узнаем, какой список нужен
	limit, ok := GetQueryINTVarFromGET(r, "limit")
	if !ok {
		limit = 100
	}
	offset, ok := GetQueryINTVarFromGET(r, "offset")
	if !ok || offset < 0 {
		offset = 0
	}
	search, ok := GetQuerySTRVarFromGET(r, "search")
	if !ok {
		search = ""
	}
	products, _ := GetQuerySTRVarFromGET(r, "products")

	productCards, total, err := shop.GetProductCardList(offset, limit, search, products == "true")
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

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	cardId, err := GetUINTVarFromRequest(r, "cardId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID группы"))
		return
	}

	// var input interface{}
	// var input map[string]interface{}
	var input = struct {
		models.ProductCard
	}{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	if input.SwitchProducts == nil {
		val := pq.StringArray{}
		input.SwitchProducts = val
	}

	card, err := account.UpdateProductCard(cardId, input.ProductCard)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Product Card Update")
	resp["card"] = card
	u.Respond(w, resp)
}

func ProductCardDelete(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	cardId, err := GetUINTVarFromRequest(r, "cardId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}


	if err = account.DeleteProductCard(cardId); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении карточки товара"))
		return
	}

	resp := u.Message(true, "DELETE Product Card Successful")
	u.Respond(w, resp)
}

/////////////////////////////////////

func ProductCreate(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
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
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания продукта"}))
		return
	}

	resp := u.Message(true, "POST Product Created")
	resp["product"] = *product
	u.Respond(w, resp)
}

func ProductListPaginationGet(w http.ResponseWriter, r *http.Request) {

	var account *models.Account
	var err error

	if isApiRequest(r) {
		account, err = GetWorkAccount(w,r)
		if err != nil || account == nil {
			return
		}
	} else {
		account, err = GetWorkAccountCheckHashId(w,r)
		if err != nil || account == nil {
			return
		}
	}

	// 2. Узнаем, какой список нужен
	limit, ok := GetQueryINTVarFromGET(r, "limit")
	if !ok {
		limit = 100
	}
	offset, ok := GetQueryINTVarFromGET(r, "offset")
	if !ok || offset < 0 {
		offset = 0
	}
	search, ok := GetQuerySTRVarFromGET(r, "search")
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

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productId, err := GetUINTVarFromRequest(r, "productId")
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

	//card, err := account.UpdateProduct(productId, input.Product)
	card, err := account.UpdateProduct(productId, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Product Update")
	resp["product"] = card
	u.Respond(w, resp)
}

func ProductDelete(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productId, err := GetUINTVarFromRequest(r, "productId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID товара"))
		return
	}


	if err = account.DeleteProduct(productId); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении карточки товара"))
		return
	}

	resp := u.Message(true, "DELETE Product Successful")
	u.Respond(w, resp)
}

////////////////////////////////////

func ProductAttributeList(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccountCheckHashId(w,r)
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