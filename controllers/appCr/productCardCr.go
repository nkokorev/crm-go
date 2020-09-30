package appCr

import (
	"encoding/json"
	"fmt"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func ProductCardCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil {
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

	productCard, err := account.CreateEntity(&input.ProductCard)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания"}))
		return
	}

	resp := u.Message(true, "POST ProductCard Created")
	resp["product_card"] = productCard
	u.Respond(w, resp)
}

func ProductCardGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	productCardId, err := utilsCr.GetUINTVarFromRequest(r, "productCardId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке productCard Id"))
		return
	}

	var productCard models.ProductCard
	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")
	// 2. Узнаем, какой id учитывается нужен
	publicOk := utilsCr.GetQueryBoolVarFromGET(r, "public_id")

	if publicOk  {
		err = account.LoadEntityByPublicId(&productCard, productCardId,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
			return
		}
	} else {
		err = account.LoadEntity(&productCard, productCardId,preloads)
		if err != nil {
			fmt.Println(err)
			u.Respond(w, u.MessageError(err, "Не удалось загрузить магазин"))
			return
		}
	}

	resp := u.Message(true, "GET ProductCard")
	resp["product_card"] = productCard
	u.Respond(w, resp)
}

func ProductCardListPaginationGet(w http.ResponseWriter, r *http.Request) {

	var account *models.Account
	var err error
	// 1. Получаем рабочий аккаунт в зависимости от источника (автома. сверка с {hashId}.)

	account, err = utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
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
		sortBy = "id"
	}
	if sortDesc {
		sortBy += " desc"
	}
	search, ok := utilsCr.GetQuerySTRVarFromGET(r, "search")
	if !ok {
		search = ""
	}
	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	// 2. Узнаем, какой список нужен
	all := utilsCr.GetQueryBoolVarFromGET(r, "all")

	// Узнаем нужен ли фильтр
	filter := map[string]interface{}{}
	webSiteId, _filterWebSite := utilsCr.GetQueryUINTVarFromGET(r, "webSiteId")
	if _filterWebSite {
		filter["web_site_id"] = webSiteId
	}

	var total int64 = 0
	productCards := make([]models.Entity,0)

	if all && len(filter) < 1{
		productCards, total, err = account.GetListEntity(&models.ProductCard{}, sortBy,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список страниц"))
			return
		}
	} else {
		// webHooks, total, err = account.GetWebHooksPaginationList(offset, limit, search)
		productCards, total, err = account.GetPaginationListEntity(&models.ProductCard{}, offset, limit, sortBy, search, filter,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список страниц"))
			return
		}
	}
	
	resp := u.Message(true, "GET product Cards PaginationList")
	resp["product_cards"] = productCards
	resp["total"] = total
	u.Respond(w, resp)
}

func ProductCardUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productCardId, err := utilsCr.GetUINTVarFromRequest(r, "productCardId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}
	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")
	var productCard models.ProductCard

	err = account.LoadEntity(&productCard, productCardId,nil)

	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось загрузить данные"))
		return
	}

	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// fix variables
	/*if err := u.ConvertMapVarsToUINT(&input, []string{"public_id"}); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}*/

	// productCard, err := account.UpdateProductCard(productCardId, &input.ProductCard)
	err = account.UpdateEntity(&productCard, input,preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH ProductCard Update")
	resp["product_card"] = productCard
	u.Respond(w, resp)
}

func ProductCardDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productCardId, err := utilsCr.GetUINTVarFromRequest(r, "productCardId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	var productCard models.ProductCard
	err = account.LoadEntity(&productCard, productCardId,nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить магазин"))
		return
	}

	if err = account.DeleteEntity(&productCard); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении карточки товара"))
		return
	}

	resp := u.Message(true, "DELETE ProductCard Successful")
	u.Respond(w, resp)
}

func ProductCardSyncProducts(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productCardId, err := utilsCr.GetUINTVarFromRequest(r, "productCardId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id productCardId"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	productCard := models.ProductCard{}
	if err =account.LoadEntity(&productCard, productCardId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Карточка товара"}))
		return
	}

	var input struct{
		Products []models.Product `json:"products"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе 1"))
		return
	}

	if err = productCard.SyncProductByIds(input.Products); err !=nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе 2"))
		return
	}

/*	_productCard := models.ProductCard{}
	if err = account.LoadEntity(&_productCard, productCardId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Карточка товара"}))
		return
	}*/

	resp := u.Message(true, "PATCH Product Card MassUpdates")
	resp["product_card"] = productCard
	u.Respond(w, resp)
}

func ProductCardRemoveProduct(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productCardId, err := utilsCr.GetUINTVarFromRequest(r, "productCardId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id emailQueueId"))
		return
	}

	productId, err := utilsCr.GetUINTVarFromRequest(r, "productId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке productId"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var productCard models.ProductCard
	if err =account.LoadEntity(&productCard, productCardId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке карточки товара"}))
		return
	}

	product := models.Product{}
	if err =account.LoadEntity(&product, productId, nil); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	if err = productCard.RemoveProduct(&product); err !=nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Ошибка удаления продукта из карточки товара"))
		return
	}

	// preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	/*filter := make(map[string]interface{},0)
	filter["product_card_id"] = productCardId*/

	/*var _productCard models.ProductCard
	if err = account.LoadEntity(&_productCard, productCardId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка загрузки карточки товара"}))
		return
	}*/

	resp := u.Message(true, "PATCH ProductCard Products Remove")
	resp["product_card"] = productCard
	u.Respond(w, resp)
}

func ProductCardAppendProduct(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productCardId, err := utilsCr.GetUINTVarFromRequest(r, "productCardId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id emailQueueId"))
		return
	}

	productId, err := utilsCr.GetUINTVarFromRequest(r, "productId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке productId"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var productCard models.ProductCard
	if err =account.LoadEntity(&productCard, productCardId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке карточки товара"}))
		return
	}

	product := models.Product{}
	if err =account.LoadEntity(&product, productId, nil); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	if err = productCard.AppendProduct(&product); err !=nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Ошибка удаления продукта из карточки товара"))
		return
	}

	var _productCard models.ProductCard
	if err = account.LoadEntity(&_productCard, productCardId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка загрузки карточки товара"}))
		return
	}

	resp := u.Message(true, "PATCH ProductCard Append Product")
	resp["product_card"] = _productCard
	u.Respond(w, resp)
}
