package appCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

// Обсерверы являются чисто системными, их нельзя добавлять или менять из-под других аккаунтов

func ObserverItemCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	if !account.IsMainAccount() {
		u.Respond(w, u.MessageError(u.Error{Message:"У вас нет прав на создание/изменение объектов этого типа"}))
		return
	}

	// Get JSON-request
	var input struct{
		models.HandlerItem
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	obItem, err := account.CreateEntity(&input.HandlerItem)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время объекта"}))
		return
	}

	resp := u.Message(true, "POST Observer Item Created")
	resp["observerItem"] = obItem
	u.Respond(w, resp)
}

func ObserverItemGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	observerItemId, err := utilsCr.GetUINTVarFromRequest(r, "observerItemId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке observer Item Id"))
		return
	}

	var obItem models.HandlerItem
	err = account.LoadEntity(&obItem, observerItemId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить observer item"))
		return
	}

	resp := u.Message(true, "GET Observer List")
	resp["observerItem"] = obItem
	u.Respond(w, resp)
}

func ObserverItemGetList(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w, r)
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
	observerItems := make([]models.Entity,0)
	if all == "true" && allOk {
		observerItems, total, err = account.GetListEntity(&models.HandlerItem{}, sortBy)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
			return
		}
	} else {
		observerItems, total, err = account.GetPaginationListEntity(&models.HandlerItem{}, offset, limit, sortBy, search)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	}



	resp := u.Message(true, "GET System Observers Pagination List")
	resp["total"] = total
	resp["observerItems"] = observerItems
	u.Respond(w, resp)
}

func ObserverItemGetListPagination(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w, r)
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
	observerItems := make([]models.Entity,0)

	if all == "true" && allOk {
		observerItems, total, err = account.GetListEntity(&models.HandlerItem{}, sortBy)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
			return
		}
	} else {
		observerItems, total, err = account.GetPaginationListEntity(&models.HandlerItem{}, offset, limit, sortBy, search)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	}



	resp := u.Message(true, "GET System Observers Pagination List")
	resp["total"] = total
	resp["observerItems"] = observerItems
	u.Respond(w, resp)
}

func ObserverItemUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	if !account.IsMainAccount() {
		u.Respond(w, u.MessageError(u.Error{Message:"У вас нет прав на создание/изменение объектов этого типа"}))
		return
	}

	observerItemId, err := utilsCr.GetUINTVarFromRequest(r, "observerItemId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке observer Item Id"))
		return
	}

	var obItem models.HandlerItem
	err = account.LoadEntity(&obItem, observerItemId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
		return
	}

	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	err = account.UpdateEntity(&obItem, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Observer Item Update")
	resp["observerItem"] = obItem
	u.Respond(w, resp)
}

func ObserverItemDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	if !account.IsMainAccount() {
		u.Respond(w, u.MessageError(u.Error{Message:"У вас нет прав на создание/изменение объектов этого типа"}))
		return
	}

	observerItemId, err := utilsCr.GetUINTVarFromRequest(r, "observerItemId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID шаблона"))
		return
	}

	var obItem models.HandlerItem
	err = account.LoadEntity(&obItem, observerItemId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
		return
	}
	if err = account.DeleteEntity(&obItem); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении"))
		return
	}

	resp := u.Message(true, "DELETE HandlerItem Successful")
	u.Respond(w, resp)
}