package appCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

// Обсерверы являются чисто системными, их нельзя добавлять или менять из-под других аккаунтов

func HandlerItemCreate(w http.ResponseWriter, r *http.Request) {

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

	handlerItem, err := account.CreateEntity(&input.HandlerItem)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время объекта"}))
		return
	}

	resp := u.Message(true, "POST HandlerItem Created")
	resp["handlerItem"] = handlerItem
	u.Respond(w, resp)
}

func HandlerItemGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	handlerItemId, err := utilsCr.GetUINTVarFromRequest(r, "handlerItemId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке handlerItem Item Id"))
		return
	}

	var handlerItem models.HandlerItem
	err = account.LoadEntity(&handlerItem, handlerItemId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить handlerItem item"))
		return
	}

	resp := u.Message(true, "GET HandlerItem List")
	resp["handlerItem"] = handlerItem
	u.Respond(w, resp)
}

func HandlerItemGetList(w http.ResponseWriter, r *http.Request) {

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
	handlerItems := make([]models.Entity,0)
	if all == "true" && allOk {
		handlerItems, total, err = account.GetListEntity(&models.HandlerItem{}, sortBy)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	} else {
		handlerItems, total, err = account.GetPaginationListEntity(&models.HandlerItem{}, offset, limit, sortBy, search, nil)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	}



	resp := u.Message(true, "GET System Observers Pagination List")
	resp["total"] = total
	resp["handlerItems"] = handlerItems
	u.Respond(w, resp)
}

func HandlerItemGetListPagination(w http.ResponseWriter, r *http.Request) {

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
	handlerItems := make([]models.Entity,0)

	if all == "true" && allOk {
		handlerItems, total, err = account.GetListEntity(&models.HandlerItem{}, sortBy)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	} else {
		handlerItems, total, err = account.GetPaginationListEntity(&models.HandlerItem{}, offset, limit, sortBy, search, nil)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	}



	resp := u.Message(true, "GET System Observers Pagination List")
	resp["total"] = total
	resp["handlerItems"] = handlerItems
	u.Respond(w, resp)
}

func HandlerItemUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	if !account.IsMainAccount() {
		u.Respond(w, u.MessageError(u.Error{Message:"У вас нет прав на создание/изменение объектов этого типа"}))
		return
	}

	handlerItemId, err := utilsCr.GetUINTVarFromRequest(r, "handlerItemId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке handlerItem Item Id"))
		return
	}

	var handlerItem models.HandlerItem
	err = account.LoadEntity(&handlerItem, handlerItemId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	err = account.UpdateEntity(&handlerItem, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH HandlerItem Update")
	resp["handlerItem"] = handlerItem
	u.Respond(w, resp)
}

func HandlerItemDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	if !account.IsMainAccount() {
		u.Respond(w, u.MessageError(u.Error{Message:"У вас нет прав на создание/изменение объектов этого типа"}))
		return
	}

	handlerItemId, err := utilsCr.GetUINTVarFromRequest(r, "handlerItemId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	var handlerItem models.HandlerItem
	err = account.LoadEntity(&handlerItem, handlerItemId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
		return
	}
	if err = account.DeleteEntity(&handlerItem); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении"))
		return
	}

	resp := u.Message(true, "DELETE HandlerItem Successful")
	u.Respond(w, resp)
}