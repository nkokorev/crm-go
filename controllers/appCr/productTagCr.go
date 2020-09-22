package appCr

import (
	"encoding/json"
	"fmt"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func ProductTagCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	var input struct{
		models.ProductTag
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	productTag, err := account.CreateEntity(&input.ProductTag)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания"}))
		return
	}

	resp := u.Message(true, "POST ProductTag Created")
	resp["product_tag"] = productTag
	u.Respond(w, resp)
}

func ProductTagGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	productTagId, err := utilsCr.GetUINTVarFromRequest(r, "productTagId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке productTag Id"))
		return
	}

	var productTag models.ProductTag
	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")
	// 2. Узнаем, какой id учитывается нужен
	publicOk := utilsCr.GetQueryBoolVarFromGET(r, "public_id")

	if publicOk  {
		err = account.LoadEntityByPublicId(&productTag, productTagId,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
			return
		}
	} else {
		err = account.LoadEntity(&productTag, productTagId,preloads)
		if err != nil {
			fmt.Println(err)
			u.Respond(w, u.MessageError(err, "Не удалось загрузить магазин"))
			return
		}
	}

	resp := u.Message(true, "GET ProductTag")
	resp["product_tag"] = productTag
	u.Respond(w, resp)
}

func ProductTagListPaginationGet(w http.ResponseWriter, r *http.Request) {

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
	productTags := make([]models.Entity,0)

	if all && len(filter) < 1{
		productTags, total, err = account.GetListEntity(&models.ProductTag{}, sortBy,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список страниц"))
			return
		}
	} else {
		// webHooks, total, err = account.GetWebHooksPaginationList(offset, limit, search)
		productTags, total, err = account.GetPaginationListEntity(&models.ProductTag{}, offset, limit, sortBy, search, filter,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список страниц"))
			return
		}
	}
	
	resp := u.Message(true, "GET product Cards PaginationList")
	resp["product_tags"] = productTags
	resp["total"] = total
	u.Respond(w, resp)
}

func ProductTagUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productTagId, err := utilsCr.GetUINTVarFromRequest(r, "productTagId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}
	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")
	var productTag models.ProductTag

	err = account.LoadEntity(&productTag, productTagId,nil)

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

	// productTag, err := account.UpdateProductTag(productTagId, &input.ProductTag)
	err = account.UpdateEntity(&productTag, input,preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH ProductTag Update")
	resp["product_tag"] = productTag
	u.Respond(w, resp)
}

func ProductTagDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productTagId, err := utilsCr.GetUINTVarFromRequest(r, "productTagId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	var productTag models.ProductTag
	err = account.LoadEntity(&productTag, productTagId,nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить магазин"))
		return
	}

	if err = account.DeleteEntity(&productTag); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении категории товара"))
		return
	}

	resp := u.Message(true, "DELETE ProductTag Successful")
	u.Respond(w, resp)
}

func ProductTagProductListPaginationGet(w http.ResponseWriter, r *http.Request) {

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

	// Узнаем нужен ли фильтр
	filter := map[string]interface{}{}
	productTagId, err := utilsCr.GetUINTVarFromRequest(r, "productTagId")
	if err == nil {
		filter["product_tag_id"] = productTagId
	}

	var productTag models.ProductTag
	err = account.LoadEntity(&productTag, productTagId,nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить магазин"))
		return
	}

	var total int64 = 0
	products := make([]models.Entity,0)

	products, total, err = productTag.GetProductPaginationList(account.Id, offset, limit, sortBy, search, filter, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список товаров"))
		return
	}

	resp := u.Message(true, "GET Product by Product Tag Pagination List")
	resp["products"] = products
	resp["total"] = total
	u.Respond(w, resp)
}

func ProductTagRemoveProduct(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productTagId, err := utilsCr.GetUINTVarFromRequest(r, "productTagId")
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

	var productTag models.ProductTag
	if err =account.LoadEntity(&productTag, productTagId, nil); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке категории товара"}))
		return
	}

	product := models.Product{}
	if err =account.LoadEntity(&product, productId, nil); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	if err = productTag.RemoveProduct(&product); err !=nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Ошибка удаления продукта из категории товара"))
		return
	}

	if err = account.LoadEntity(&productTag, productTagId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке категории товара"}))
		return
	}

	resp := u.Message(true, "PATCH ProductTag Products Remove")
	resp["product_tag"] = productTag
	u.Respond(w, resp)
}

func ProductTagAppendProduct(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productTagId, err := utilsCr.GetUINTVarFromRequest(r, "productTagId")
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

	var productTag models.ProductTag
	if err =account.LoadEntity(&productTag, productTagId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке категории товара"}))
		return
	}

	product := models.Product{}
	if err =account.LoadEntity(&product, productId, nil); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	if err = productTag.AppendProduct(&product); err !=nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Ошибка удаления продукта из категории товара"))
		return
	}

	var _productTag models.ProductTag
	if err = account.LoadEntity(&_productTag, productTagId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка загрузки категории товара"}))
		return
	}

	resp := u.Message(true, "PATCH ProductTag Append Product")
	resp["product_tag"] = _productTag
	u.Respond(w, resp)
}
