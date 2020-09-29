package appCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func EventCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	var input struct{
		models.Event
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	event, err := account.CreateEntity(&input.Event)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время объекта"}))
		return
	}

	resp := u.Message(true, "POST Event Item Created")
	resp["event"] = event
	u.Respond(w, resp)
}

func EventGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	eventId, err := utilsCr.GetUINTVarFromRequest(r, "eventId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке observer Item Id"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var event models.Event
	err = account.LoadEntity(&event, eventId, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить observer item"))
		return
	}

	resp := u.Message(true, "GET Event List")
	resp["event"] = event
	u.Respond(w, resp)
}

func EventGetListPagination(w http.ResponseWriter, r *http.Request) {

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
	events := make([]models.Entity,0)

	if all {
		events, total, err = account.GetListEntity(&models.Event{}, sortBy,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	} else {
		events, total, err = account.GetPaginationListEntity(&models.Event{}, offset, limit, sortBy, search, nil,nil)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	}



	resp := u.Message(true, "GET System Events List")
	resp["total"] = total
	resp["events"] = events
	u.Respond(w, resp)
}

func EventUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	eventId, err := utilsCr.GetUINTVarFromRequest(r, "eventId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке observer Item Id"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var event models.Event
	err = account.LoadEntity(&event, eventId, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	// Проверка на права изменения
	if event.AccountId != account.Id {
		u.Respond(w, u.MessageError(u.Error{ Message:"У вас нет прав на создание/изменение объектов этого типа"}))
		return
	}

	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	err = account.UpdateEntity(&event, input, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Event Item Update")
	resp["event"] = event
	u.Respond(w, resp)
}

func EventDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	eventId, err := utilsCr.GetUINTVarFromRequest(r, "eventId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	var event models.Event
	err = account.LoadEntity(&event, eventId, nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
		return
	}

	// Проверка на права изменения
	if event.AccountId != account.Id {
		u.Respond(w, u.MessageError(u.Error{ Message:"У вас нет прав на создание/изменение объектов этого типа"}))
		return
	}

	// Удаляем объект
	if err = account.DeleteEntity(&event); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении"))
		return
	}

	resp := u.Message(true, "DELETE Event Successful")
	u.Respond(w, resp)
}

func EventExecute(w http.ResponseWriter, r *http.Request) {


	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	eventId, err := utilsCr.GetUINTVarFromRequest(r, "eventId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке {eventId}"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var event models.Event
	err = account.LoadEntity(&event, eventId, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти событие. Проверьте /.../{eventId}"))
		return
	}

	// Проверяем, что для этого событие разрешен вызов по API
	if !event.AvailableAPI {
		u.Respond(w, u.MessageError(err, "Вызов не удался: запрещен вызов события по API"))
		return
	}

	// Определяем тип запроса
	if r.Method == http.MethodGet {
		
	}
	
	// Собираем входящие данные, если запрос не GET
	var input struct{
		RecipientListIds string `json:"recipient_list_ids"`
		Payload map[string]interface{} `json:"payload"` // JSONData for
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// Устанавливаем контекстные данные для события
	event.SetPayload(input.Payload)
	// event.Payload = input.Payload
	

	// event.Ex


	resp := u.Message(true, "GET Event Execute")
	u.Respond(w, resp)
}