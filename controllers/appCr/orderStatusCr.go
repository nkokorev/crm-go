package appCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func OrderStatusCreate(w http.ResponseWriter, r *http.Request) {

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
		models.OrderStatus
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	orderStatus, err := account.CreateEntity(&input.OrderStatus)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания статуса"}))
		return
	}

	resp := u.Message(true, "POST OrderStatus Created")
	resp["order_status"] = orderStatus
	u.Respond(w, resp)
}

func OrderStatusGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	orderStatusId, err := utilsCr.GetUINTVarFromRequest(r, "orderStatusId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке web site Id"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var orderStatus models.OrderStatus
	err = account.LoadEntity(&orderStatus, orderStatusId,preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	resp := u.Message(true, "GET OrderStatus")
	resp["order_status"] = orderStatus
	u.Respond(w, resp)
}

func OrderStatusGetList(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w, r)
	if err != nil || account == nil {
		return
	}

	var total int64 = 0
	orderStatuses := make([]models.Entity,0)

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	orderStatuses, total, err = account.GetListEntity(&models.OrderStatus{},"id",preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	resp := u.Message(true, "GET OrderStatus List")
	resp["total"] = total
	resp["order_statuses"] = orderStatuses
	u.Respond(w, resp)
}

func OrderStatusUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	if !account.IsMainAccount() {
		u.Respond(w, u.MessageError(u.Error{Message:"У вас нет прав на создание/изменение объектов этого типа"}))
		return
	}

	orderStatusId, err := utilsCr.GetUINTVarFromRequest(r, "orderStatusId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var orderStatus models.OrderStatus
	err = account.LoadEntity(&orderStatus, orderStatusId, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	err = account.UpdateEntity(&orderStatus, input, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH OrderStatus Update")
	resp["order_status"] = orderStatus
	u.Respond(w, resp)
}

func OrderStatusDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	if !account.IsMainAccount() {
		u.Respond(w, u.MessageError(u.Error{Message:"У вас нет прав на создание/изменение объектов этого типа"}))
		return
	}

	orderStatusId, err := utilsCr.GetUINTVarFromRequest(r, "orderStatusId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	var orderStatus models.OrderStatus
	err = account.LoadEntity(&orderStatus, orderStatusId, nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось загрузить данные"))
		return
	}
	if err = account.DeleteEntity(&orderStatus); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении"))
		return
	}

	resp := u.Message(true, "DELETE OrderStatus Successful")
	u.Respond(w, resp)
}
