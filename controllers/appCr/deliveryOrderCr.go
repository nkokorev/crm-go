package appCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func DeliveryOrderCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	var input struct {
		models.DeliveryOrder
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	deliveryOrder, err := account.CreateEntity(&input.DeliveryOrder)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания заявки"}))
		return
	}

	resp := u.Message(true, "POST DeliveryOrder Created")
	resp["deliveryOrder"] = deliveryOrder
	u.Respond(w, resp)
}

func DeliveryOrderGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	deliveryOrderId, err := utilsCr.GetUINTVarFromRequest(r, "deliveryOrderId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке web site Id"))
		return
	}

	// Public ID!!
	var deliveryOrder models.DeliveryOrder
	err = account.LoadEntityByPublicId(&deliveryOrder, deliveryOrderId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	resp := u.Message(true, "GET DeliveryOrder")
	resp["deliveryOrder"] = deliveryOrder
	u.Respond(w, resp)
}

func DeliveryOrderGetListPagination(w http.ResponseWriter, r *http.Request) {
	
	account, err := utilsCr.GetWorkAccount(w, r)
	if err != nil || account == nil {
		return
	}

	limit, ok := utilsCr.GetQueryINTVarFromGET(r, "limit")
	if !ok {
		limit = 25
	}
	if limit > 100 { limit = 100 }
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

	var total uint = 0
	deliveryOrders := make([]models.Entity,0)
	
	deliveryOrders, total, err = account.GetPaginationListEntity(&models.DeliveryOrder{}, offset, limit, sortBy, search)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	resp := u.Message(true, "GET DeliveryOrder Pagination List")
	resp["total"] = total
	resp["deliveryOrders"] = deliveryOrders
	u.Respond(w, resp)
}

func DeliveryOrderUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	deliveryOrderId, err := utilsCr.GetUINTVarFromRequest(r, "deliveryOrderId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	var deliveryOrder models.DeliveryOrder
	err = account.LoadEntity(&deliveryOrder, deliveryOrderId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	err = account.UpdateEntity(&deliveryOrder, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH DeliveryOrder Update")
	resp["deliveryOrder"] = deliveryOrder
	u.Respond(w, resp)
}

func DeliveryOrderDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	deliveryOrderId, err := utilsCr.GetUINTVarFromRequest(r, "deliveryOrderId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	var deliveryOrder models.DeliveryOrder
	err = account.LoadEntity(&deliveryOrder, deliveryOrderId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
		return
	}
	if err = account.DeleteEntity(&deliveryOrder); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении магазина"))
		return
	}

	resp := u.Message(true, "DELETE DeliveryOrder Successful")
	u.Respond(w, resp)
}
