package appCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"

)

func WebHookCreate(w http.ResponseWriter, r *http.Request) {

	var account *models.Account
	var err error
	account, err = utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	var input struct{
		models.WebHook
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	entity, err := account.CreateEntity(&input.WebHook)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания WebHook"}))
		return
	}

	resp := u.Message(true, "POST WebHook Create")
	resp["webHook"] = entity
	u.Respond(w, resp)
}

func WebHookGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	webHookId, err := utilsCr.GetUINTVarFromRequest(r, "webHookId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id webHook"))
		return
	}

	var webHook models.WebHook
	// 2. Узнаем, какой id учитывается нужен
	publicOk := utilsCr.GetQueryBoolVarFromGET(r, "publicId")

	if publicOk  {
		err = account.LoadEntityByPublicId(&webHook, webHookId)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
			return
		}
	} else {
		err = account.LoadEntity(&webHook, webHookId)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
			return
		}

	}


	resp := u.Message(true, "GET Web Hook")
	resp["webHook"] = webHook
	u.Respond(w, resp)
}

func WebHookExecute(w http.ResponseWriter, r *http.Request) {

	var account *models.Account
	var err error

	account, err = utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	webHookId, err := utilsCr.GetUINTVarFromRequest(r, "webHookId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id webHook"))
		return
	}

	var webHook models.WebHook
	err = account.LoadEntity(&webHook, webHookId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти вебхук"))
		return
	}

	go webHook.Execute(nil)

	resp := u.Message(true, "GET Web Hook Call")
	u.Respond(w, resp)
}

func WebHookListPaginationGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
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
	webHooks := make([]models.Entity,0)

	if all == "true" && allOk {
		webHooks, total, err = account.GetListEntity(&models.WebHook{}, sortBy)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список ВебХуков"))
			return
		}
	} else {
		// webHooks, total, err = account.GetWebHooksPaginationList(offset, limit, search)
		webHooks, total, err = account.GetPaginationListEntity(&models.WebHook{}, offset, limit, sortBy, search, nil)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список ВебХуков"))
			return
		}
	}


	resp := u.Message(true, "GET WebHooks PaginationList")
	resp["webHooks"] = webHooks
	resp["total"] = total
	u.Respond(w, resp)
}

func WebHookUpdate(w http.ResponseWriter, r *http.Request) {

	var account *models.Account
	var err error
	account, err = utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	webHookId, err := utilsCr.GetUINTVarFromRequest(r, "webHookId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id группы"))
		return
	}

	// var input interface{}
	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	var webHook models.WebHook
	if err = account.LoadEntity(&webHook, webHookId); err != nil {
		u.Respond(w, u.MessageError(err, "WEbHook не найден"))
		return
	}

	// webHook, err := account.UpdateWebHook(webHookId, input)
	if err = account.UpdateEntity(&webHook, input); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH WebHook Update")
	resp["webHook"] = webHook
	u.Respond(w, resp)
}

func WebHookDelete(w http.ResponseWriter, r *http.Request) {

	var account *models.Account
	var err error
	account, err = utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	webHookId, err := utilsCr.GetUINTVarFromRequest(r, "webHookId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id группы"))
		return
	}

	var webHook models.WebHook
	if err = account.LoadEntity(&webHook, webHookId); err != nil {
		u.Respond(w, u.MessageError(err, "WEbHook не найден"))
		return
	}

	// webHook, err := account.UpdateWebHook(webHookId, input)
	if err = account.DeleteEntity(&webHook); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении"))
		return
	}

	resp := u.Message(true, "DELETE WebHook Successful")
	u.Respond(w, resp)
}
