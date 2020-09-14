package appCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func PaymentModeCreate(w http.ResponseWriter, r *http.Request) {

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
		models.PaymentMode
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	paymentMode, err := account.CreateEntity(&input.PaymentMode)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания"}))
		return
	}

	resp := u.Message(true, "POST PaymentMode Created")
	resp["payment_mode"] = paymentMode
	u.Respond(w, resp)
}

func PaymentModeGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	paymentModeId, err := utilsCr.GetUINTVarFromRequest(r, "paymentModeId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке paymentModeId"))
		return
	}
	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var paymentMode models.PaymentMode
	err = account.LoadEntity(&paymentMode, paymentModeId,preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	resp := u.Message(true, "GET PaymentMode")
	resp["payment_mode"] = paymentMode
	u.Respond(w, resp)
}

func PaymentModeGetList(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w, r)
	if err != nil || account == nil {
		return
	}

	var total int64 = 0
	paymentModes := make([]models.Entity,0)
	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")
	paymentModes, total, err = account.GetListEntity(&models.PaymentMode{},"id",preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	resp := u.Message(true, "GET PaymentMode List")
	resp["total"] = total
	resp["payment_modes"] = paymentModes
	u.Respond(w, resp)
}

func PaymentModeUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	if !account.IsMainAccount() {
		u.Respond(w, u.MessageError(u.Error{Message:"У вас нет прав на создание/изменение объектов этого типа"}))
		return
	}

	paymentModeId, err := utilsCr.GetUINTVarFromRequest(r, "paymentModeId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}
	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")
	var paymentMode models.PaymentMode
	err = account.LoadEntity(&paymentMode, paymentModeId,preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	err = account.UpdateEntity(&paymentMode, input,preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH PaymentMode Update")
	resp["payment_mode"] = paymentMode
	u.Respond(w, resp)
}

func PaymentModeDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	if !account.IsMainAccount() {
		u.Respond(w, u.MessageError(u.Error{Message:"У вас нет прав на создание/изменение объектов этого типа"}))
		return
	}

	paymentModeId, err := utilsCr.GetUINTVarFromRequest(r, "paymentModeId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	var paymentMode models.PaymentMode
	err = account.LoadEntity(&paymentMode, paymentModeId,nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось загрузить данные"))
		return
	}
	if err = account.DeleteEntity(&paymentMode); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении способа оплаты"))
		return
	}

	resp := u.Message(true, "DELETE PaymentMode Successful")
	u.Respond(w, resp)
}
