package appCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func EventListenerCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	var input struct{
		models.EventListener
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	eventListener, err := account.CreateEntity(&input.EventListener)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания слушателя"}))
		return
	}

	resp := u.Message(true, "POST Event Listener Created")
	resp["event_listener"] = eventListener
	u.Respond(w, resp)
}

func EventListenerGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	eventListenerId, err := utilsCr.GetUINTVarFromRequest(r, "eventListenerId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке web site Id"))
		return
	}
	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var eventListener models.EventListener
	err = account.LoadEntity(&eventListener, eventListenerId, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	resp := u.Message(true, "GET Event Listener")
	resp["event_listener"] = eventListener
	u.Respond(w, resp)
}

func EventListenerGetListPagination(w http.ResponseWriter, r *http.Request) {

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
		sortBy = "id"
	}
	if sortDesc {
		sortBy += " desc"
	}

	search, ok := utilsCr.GetQuerySTRVarFromGET(r, "search")
	if !ok {
		search = ""
	}
	all := utilsCr.GetQueryBoolVarFromGET(r, "all")

	var total int64 = 0
	eventListeners := make([]models.Entity,0)

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	if all {
		eventListeners, total, err = account.GetListEntity(&models.EventListener{}, sortBy,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	} else {
		eventListeners, total, err = account.GetPaginationListEntity(&models.EventListener{}, offset, limit, sortBy, search, nil,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	}

	resp := u.Message(true, "GET Event Listener Pagination List")
	resp["total"] = total
	resp["event_listeners"] = eventListeners
	u.Respond(w, resp)
}

func EventListenerUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	eventListenerId, err := utilsCr.GetUINTVarFromRequest(r, "eventListenerId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var eventListener models.EventListener
	err = account.LoadEntity(&eventListener, eventListenerId, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	err = account.UpdateEntity(&eventListener, input, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Event Listener Update")
	resp["event_listener"] = eventListener
	u.Respond(w, resp)
}

func EventListenerDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	eventListenerId, err := utilsCr.GetUINTVarFromRequest(r, "eventListenerId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	var eventListener models.EventListener
	err = account.LoadEntity(&eventListener, eventListenerId, nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}
	if err = account.DeleteEntity(&eventListener); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении магазина"))
		return
	}

	resp := u.Message(true, "DELETE Event Listener Successful")
	u.Respond(w, resp)
}
