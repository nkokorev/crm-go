package appCr

import (
	"encoding/json"
	"fmt"
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

	wh, err := account.CreateWebHook(input.WebHook)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания WebHook"}))
		return
	}

	resp := u.Message(true, "POST WebHook Create")
	resp["webHook"] = *wh
	u.Respond(w, resp)
}

func WebHookGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	webHookId, err := utilsCr.GetUINTVarFromRequest(r, "webHookId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID webHook"))
		return
	}

	webHook, err := account.GetWebHook(webHookId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	resp := u.Message(true, "GET Web Hook")
	resp["webHook"] = webHook
	u.Respond(w, resp)
}

func WebHookCall(w http.ResponseWriter, r *http.Request) {

	var account *models.Account
	var err error

	account, err = utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	webHookId, err := utilsCr.GetUINTVarFromRequest(r, "webHookId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID webHook"))
		return
	}

	webHook, err := account.GetWebHook(webHookId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти вебхук"))
		return
	}

	go webHook.Call(nil)

	resp := u.Message(true, "GET Web Hook Call")
	u.Respond(w, resp)
}

func WebHookListPaginationGet(w http.ResponseWriter, r *http.Request) {

	var account *models.Account
	var err error
	// 1. Получаем рабочий аккаунт в зависимости от источника (автома. сверка с {hashId}.)

	account, err = utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	// 2. Узнаем, какой список нужен
	all, allOk := utilsCr.GetQuerySTRVarFromGET(r, "all")

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

	webHooks := make([]models.WebHook,0)
	total := 0

	if all == "true" && allOk {
		webHooks, err = account.GetWebHooks()
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список ВебХуков"))
			return
		}
	} else {
		webHooks, total, err = account.GetWebHooksPaginationList(offset, limit, search)
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

	webHook, err := account.UpdateWebHook(webHookId, input)
	if err != nil {
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
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID группы"))
		return
	}


	if err = account.DeleteWebHook(webHookId); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении webHook"))
		return
	}

	resp := u.Message(true, "DELETE WebHook Successful")
	u.Respond(w, resp)
}
