package appCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func DeliveryStatusCreate(w http.ResponseWriter, r *http.Request) {

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
		models.DeliveryStatus
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	deliveryStatus, err := account.CreateEntity(&input.DeliveryStatus)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания ключа"}))
		return
	}

	resp := u.Message(true, "POST DeliveryStatus Created")
	resp["deliveryStatus"] = deliveryStatus
	u.Respond(w, resp)
}

func DeliveryStatusGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	deliveryStatusId, err := utilsCr.GetUINTVarFromRequest(r, "deliveryStatusId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке web site Id"))
		return
	}

	var deliveryStatus models.DeliveryStatus
	err = account.LoadEntity(&deliveryStatus, deliveryStatusId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	resp := u.Message(true, "GET DeliveryStatus")
	resp["deliveryStatus"] = deliveryStatus
	u.Respond(w, resp)
}

func DeliveryStatusGetList(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w, r)
	if err != nil || account == nil {
		return
	}

	var total uint = 0
	deliveryStatuses := make([]models.Entity,0)
	
	deliveryStatuses, total, err = account.GetListEntity(&models.DeliveryStatus{},"id")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	resp := u.Message(true, "GET DeliveryStatus List")
	resp["total"] = total
	resp["deliveryStatuses"] = deliveryStatuses
	u.Respond(w, resp)
}

func DeliveryStatusUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	if !account.IsMainAccount() {
		u.Respond(w, u.MessageError(u.Error{Message:"У вас нет прав на создание/изменение объектов этого типа"}))
		return
	}

	deliveryStatusId, err := utilsCr.GetUINTVarFromRequest(r, "deliveryStatusId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	var deliveryStatus models.DeliveryStatus
	err = account.LoadEntity(&deliveryStatus, deliveryStatusId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	err = account.UpdateEntity(&deliveryStatus, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH DeliveryStatus Update")
	resp["deliveryStatus"] = deliveryStatus
	u.Respond(w, resp)
}

func DeliveryStatusDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	if !account.IsMainAccount() {
		u.Respond(w, u.MessageError(u.Error{Message:"У вас нет прав на создание/изменение объектов этого типа"}))
		return
	}

	deliveryStatusId, err := utilsCr.GetUINTVarFromRequest(r, "deliveryStatusId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	var deliveryStatus models.DeliveryStatus
	err = account.LoadEntity(&deliveryStatus, deliveryStatusId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
		return
	}
	if err = account.DeleteEntity(&deliveryStatus); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении магазина"))
		return
	}

	resp := u.Message(true, "DELETE DeliveryStatus Successful")
	u.Respond(w, resp)
}
