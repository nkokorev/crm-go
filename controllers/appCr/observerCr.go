package appCr

import (
	"encoding/json"
	"fmt"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func ObserverCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
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

	shopE, err := account.CreateEntity(&input.Shop)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания ключа"}))
		return
	}
	shop, ok := shopE.(*models.Shop)
	if !ok {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка приведения типов при создании магазина"}))
		return
	}

	resp := u.Message(true, "POST Shop Created")
	resp["shop"] = *shop
	u.Respond(w, resp)
}

func ObserverGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	shopId, err := utilsCr.GetUINTVarFromRequest(r, "shopId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке shop Id"))
		return
	}

	var shop models.Shop
	err = account.LoadEntity(&shop, shopId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
		return
	}

	resp := u.Message(true, "GET Shop List")
	resp["shop"] = shop
	u.Respond(w, resp)
}

func ObserverGetListPagination(w http.ResponseWriter, r *http.Request) {

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
	all, allOk := utilsCr.GetQuerySTRVarFromGET(r, "all")

	var total uint = 0
	observers := make([]models.Entity,0)


	if all == "true" && allOk {
		observers, total, err = account.GetListEntity(&models.Observer{}, sortBy)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
			return
		}
	} else {
		observers, total, err = account.GetPaginationListEntity(&models.Observer{}, offset, limit, sortBy, search)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	}

	resp := u.Message(true, "GET System Observers Pagination List")
	resp["total"] = total
	resp["observers"] = observers
	u.Respond(w, resp)
}

func ObserverUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	observerId, err := utilsCr.GetUINTVarFromRequest(r, "observerId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID шаблона"))
		return
	}

	var observer models.Observer
	err = account.LoadEntity(&observer, observerId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
		return
	}

	fmt.Println("Load Event.Name: ", observer.Event)

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	err = account.UpdateEntity(&observer, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Observer Update")
	resp["observer"] = observer
	u.Respond(w, resp)
}

func ObserverDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	shopId, err := utilsCr.GetUINTVarFromRequest(r, "shopId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID шаблона"))
		return
	}

	var shop models.Shop
	err = account.LoadEntity(&shop, shopId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
		return
	}
	if err = account.DeleteEntity(&shop); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении магазина"))
		return
	}

	resp := u.Message(true, "DELETE Shop Successful")
	u.Respond(w, resp)
}

///////////////////////////



func EventSystemListGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	resp := u.Message(true, "GET System Event List")
	resp["systemEvents"] = models.GetSystemEventList()
	u.Respond(w, resp)
}

func HandlersSystemListGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	resp := u.Message(true, "GET systemHandlers Event List")
	resp["systemHandlers"] = models.GetSystemHandlerList()
	u.Respond(w, resp)
}