package appCr

import (
	"encoding/json"
	"fmt"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func EventItemCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	var input struct{
		models.EventItem
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	eventItem, err := account.CreateEntity(&input.EventItem)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время объекта"}))
		return
	}

	resp := u.Message(true, "POST Event Item Created")
	resp["event_item"] = eventItem
	u.Respond(w, resp)
}

func EventItemGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	eventItemId, err := utilsCr.GetUINTVarFromRequest(r, "eventItemId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке observer Item Id"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var eventItem models.EventItem
	err = account.LoadEntity(&eventItem, eventItemId, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить observer item"))
		return
	}

	resp := u.Message(true, "GET Event List")
	resp["event_item"] = eventItem
	u.Respond(w, resp)
}

func EventItemGetListPagination(w http.ResponseWriter, r *http.Request) {

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
	// 2. Узнаем, какой список нужен
	all := utilsCr.GetQueryBoolVarFromGET(r, "all")

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var total int64 = 0
	observerItems := make([]models.Entity,0)

	if all {
		observerItems, total, err = account.GetListEntity(&models.EventItem{}, sortBy,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	} else {
		observerItems, total, err = account.GetPaginationListEntity(&models.EventItem{}, offset, limit, sortBy, search, nil,nil)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	}



	resp := u.Message(true, "GET System Event Pagination List")
	resp["total"] = total
	resp["event_items"] = observerItems
	u.Respond(w, resp)
}

func EventItemUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	eventItemId, err := utilsCr.GetUINTVarFromRequest(r, "eventItemId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке observer Item Id"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var eventItem models.EventItem
	err = account.LoadEntity(&eventItem, eventItemId, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	// Проверка на права изменения
	if eventItem.AccountId != account.Id {
		u.Respond(w, u.MessageError(u.Error{ Message:"У вас нет прав на создание/изменение объектов этого типа"}))
		return
	}

	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	err = account.UpdateEntity(&eventItem, input, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Event Item Update")
	resp["event_item"] = eventItem
	u.Respond(w, resp)
}

func EventItemDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	eventItemId, err := utilsCr.GetUINTVarFromRequest(r, "eventItemId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	var eventItem models.EventItem
	err = account.LoadEntity(&eventItem, eventItemId, nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
		return
	}

	// Проверка на права изменения
	if eventItem.AccountId != account.Id {
		u.Respond(w, u.MessageError(u.Error{ Message:"У вас нет прав на создание/изменение объектов этого типа"}))
		return
	}

	// Удаляем объект
	if err = account.DeleteEntity(&eventItem); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении"))
		return
	}

	resp := u.Message(true, "DELETE EventItem Successful")
	u.Respond(w, resp)
}

func EventItemExecute(w http.ResponseWriter, r *http.Request) {


	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	eventItemId, err := utilsCr.GetUINTVarFromRequest(r, "eventItemId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке {eventItemId}"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var eventItem models.EventItem
	err = account.LoadEntity(&eventItem, eventItemId, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить событие"))
		return
	}

	// Проверяем, что для этого событие разрешен вызов по API
	if !eventItem.AvailableAPI {
		u.Respond(w, u.MessageError(err, "Вызов не удался: запрещен вызов события по API"))
		return
	}

	// Собираем входящие данные
	var input struct{
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	switch input.Status {
	case models.WorkStatusPending:
		err := emailCampaign.SetPendingStatus()
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	case models.WorkStatusPlanned:
		err := emailCampaign.SetPlannedStatus()
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	case models.WorkStatusActive:
		err := emailCampaign.SetActiveStatus()
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	case models.WorkStatusPaused:
		err := emailCampaign.SetPausedStatus()
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	case models.WorkStatusCompleted:
		err := emailCampaign.SetCompletedStatus()
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	case models.WorkStatusFailed:
		err := emailCampaign.SetFailedStatus(input.Reason)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	case models.WorkStatusCancelled:
		err := emailCampaign.SetCancelledStatus()
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	}

	resp := u.Message(true, "GET Email Campaign Execute")
	resp["email_campaign"] = emailCampaign
	u.Respond(w, resp)
}