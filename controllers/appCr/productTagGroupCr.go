package appCr

import (
	"encoding/json"
	"fmt"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func ProductTagGroupCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	var input struct{
		models.ProductTagGroup
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	productTagGroup, err := account.CreateEntity(&input.ProductTagGroup)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания"}))
		return
	}

	resp := u.Message(true, "POST ProductTagGroup Created")
	resp["product_tag_group"] = productTagGroup
	u.Respond(w, resp)
}

func ProductTagGroupGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	productTagGroupId, err := utilsCr.GetUINTVarFromRequest(r, "productTagGroupId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке productTagGroup Id"))
		return
	}

	var productTagGroup models.ProductTagGroup
	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")
	// 2. Узнаем, какой id учитывается нужен
	publicOk := utilsCr.GetQueryBoolVarFromGET(r, "public_id")

	if publicOk  {
		err = account.LoadEntityByPublicId(&productTagGroup, productTagGroupId,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
			return
		}
	} else {
		err = account.LoadEntity(&productTagGroup, productTagGroupId,preloads)
		if err != nil {
			fmt.Println(err)
			u.Respond(w, u.MessageError(err, "Не удалось загрузить магазин"))
			return
		}
	}

	resp := u.Message(true, "GET ProductTagGroup")
	resp["product_tag_group"] = productTagGroup
	u.Respond(w, resp)
}

func ProductTagGroupListPaginationGet(w http.ResponseWriter, r *http.Request) {

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

	var total int64 = 0
	productTagGroups := make([]models.Entity,0)

	if all && len(filter) < 1{
		productTagGroups, total, err = account.GetListEntity(&models.ProductTagGroup{}, sortBy,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список страниц"))
			return
		}
	} else {
		// webHooks, total, err = account.GetWebHooksPaginationList(offset, limit, search)
		productTagGroups, total, err = account.GetPaginationListEntity(&models.ProductTagGroup{}, offset, limit, sortBy, search, filter,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список страниц"))
			return
		}
	}
	
	resp := u.Message(true, "GET product Cards PaginationList")
	resp["product_tag_groups"] = productTagGroups
	resp["total"] = total
	u.Respond(w, resp)
}

func ProductTagGroupUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productTagGroupId, err := utilsCr.GetUINTVarFromRequest(r, "productTagGroupId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}
	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")
	var productTagGroup models.ProductTagGroup

	err = account.LoadEntity(&productTagGroup, productTagGroupId,nil)

	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось загрузить данные"))
		return
	}

	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// productTagGroup, err := account.UpdateProductTagGroup(productTagGroupId, &input.ProductTagGroup)
	err = account.UpdateEntity(&productTagGroup, input,preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH ProductTagGroup Update")
	resp["product_tag_group"] = productTagGroup
	u.Respond(w, resp)
}

func ProductTagGroupDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productTagGroupId, err := utilsCr.GetUINTVarFromRequest(r, "productTagGroupId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	var productTagGroup models.ProductTagGroup
	err = account.LoadEntity(&productTagGroup, productTagGroupId,nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить магазин"))
		return
	}

	if err = account.DeleteEntity(&productTagGroup); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении категории товара"))
		return
	}

	resp := u.Message(true, "DELETE ProductTagGroup Successful")
	u.Respond(w, resp)
}

func ProductTagGroupTagListPaginationGet(w http.ResponseWriter, r *http.Request) {

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
	productTagGroupId, err := utilsCr.GetUINTVarFromRequest(r, "productTagGroupId")
	if err == nil {
		filter["product_tag_group_id"] = productTagGroupId
	}

	var productTagGroup models.ProductTagGroup
	err = account.LoadEntity(&productTagGroup, productTagGroupId,nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить productTagGroup"))
		return
	}

	var total int64 = 0
	productTags := make([]models.Entity,0)

	productTags, total, err = productTagGroup.GetTagPaginationList(account.Id, offset, limit, sortBy, search, filter, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список товаров"))
		return
	}

	resp := u.Message(true, "GET ProductTags by Product Tag Group Pagination List")
	resp["product_tags"] = productTags
	resp["total"] = total
	u.Respond(w, resp)
}

func ProductTagGroupRemoveProductTag(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productTagGroupId, err := utilsCr.GetUINTVarFromRequest(r, "productTagGroupId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id emailQueueId"))
		return
	}

	productTagId, err := utilsCr.GetUINTVarFromRequest(r, "productTagId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке productId"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var productTagGroup models.ProductTagGroup
	if err =account.LoadEntity(&productTagGroup, productTagGroupId, nil); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке категории товара"}))
		return
	}

	productTag := models.ProductTag{}
	if err =account.LoadEntity(&productTag, productTagId, nil); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	if err = productTagGroup.RemoveProductTag(&productTag); err !=nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Ошибка удаления тега из группы"))
		return
	}

	if err = account.LoadEntity(&productTagGroup, productTagGroupId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке группы тегов"}))
		return
	}

	resp := u.Message(true, "PATCH ProductTagGroup Products Remove")
	resp["product_tag_group"] = productTagGroup
	u.Respond(w, resp)
}

func ProductTagGroupAppendProductTag(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productTagGroupId, err := utilsCr.GetUINTVarFromRequest(r, "productTagGroupId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id emailQueueId"))
		return
	}

	productTagId, err := utilsCr.GetUINTVarFromRequest(r, "productTagId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке productId"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var productTagGroup models.ProductTagGroup
	if err =account.LoadEntity(&productTagGroup, productTagGroupId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке категории товара"}))
		return
	}

	productTag := models.ProductTag{}
	if err =account.LoadEntity(&productTag, productTagId, nil); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	if err = productTagGroup.AppendProductTag(&productTag); err !=nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Ошибка добавления тега в группу"))
		return
	}

	var _productTagGroup models.ProductTagGroup
	if err = account.LoadEntity(&_productTagGroup, productTagGroupId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка загрузки категории товара"}))
		return
	}

	resp := u.Message(true, "PATCH ProductTagGroup Append Product")
	resp["product_tag_group"] = _productTagGroup
	u.Respond(w, resp)
}
