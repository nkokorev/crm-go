package appCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"

)

func QuestionCreate(w http.ResponseWriter, r *http.Request) {

	var account *models.Account
	var err error
	account, err = utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	var input struct{
		models.Question
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	entity, err := account.CreateEntity(&input.Question)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания Question"}))
		return
	}

	resp := u.Message(true, "POST Question Create")
	resp["question"] = entity
	u.Respond(w, resp)
}

func QuestionGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	questionId, err := utilsCr.GetUINTVarFromRequest(r, "questionId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id question"))
		return
	}

	var question models.Question
	// 2. Узнаем, какой id учитывается нужен
	publicOk := utilsCr.GetQueryBoolVarFromGET(r, "public_id")

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	if publicOk  {
		err = account.LoadEntityByPublicId(&question, questionId,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
			return
		}
	} else {
		err = account.LoadEntity(&question, questionId,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
			return
		}

	}


	resp := u.Message(true, "GET Question")
	resp["question"] = question
	u.Respond(w, resp)
}

func QuestionListPaginationGet(w http.ResponseWriter, r *http.Request) {

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

	var total int64 = 0
	webHooks := make([]models.Entity,0)

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	if all {
		webHooks, total, err = account.GetListEntity(&models.Question{}, sortBy,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список ВебХуков"))
			return
		}
	} else {
		// webHooks, total, err = account.GetQuestionsPaginationList(offset, limit, search)
		webHooks, total, err = account.GetPaginationListEntity(&models.Question{}, offset, limit, sortBy, search, nil,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список ВебХуков"))
			return
		}
	}


	resp := u.Message(true, "GET Questions PaginationList")
	resp["questions"] = webHooks
	resp["total"] = total
	u.Respond(w, resp)
}

func QuestionUpdate(w http.ResponseWriter, r *http.Request) {

	var account *models.Account
	var err error
	account, err = utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	questionId, err := utilsCr.GetUINTVarFromRequest(r, "questionId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id группы"))
		return
	}
	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	// var input interface{}
	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	var question models.Question
	if err = account.LoadEntity(&question, questionId,preloads); err != nil {
		u.Respond(w, u.MessageError(err, "WEbHook не найден"))
		return
	}

	// question, err := account.UpdateQuestion(questionId, input)
	if err = account.UpdateEntity(&question, input,preloads); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Question Update")
	resp["question"] = question
	u.Respond(w, resp)
}

func QuestionDelete(w http.ResponseWriter, r *http.Request) {

	var account *models.Account
	var err error
	account, err = utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	questionId, err := utilsCr.GetUINTVarFromRequest(r, "questionId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id группы"))
		return
	}

	var question models.Question
	if err = account.LoadEntity(&question, questionId,nil); err != nil {
		u.Respond(w, u.MessageError(err, "WEbHook не найден"))
		return
	}

	// question, err := account.UpdateQuestion(questionId, input)
	if err = account.DeleteEntity(&question); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении"))
		return
	}

	resp := u.Message(true, "DELETE Question Successful")
	u.Respond(w, resp)
}

func QuestionStatus(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	questionId, err := utilsCr.GetUINTVarFromRequest(r, "questionId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var question models.Question
	err = account.LoadEntity(&question, questionId, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить данные вопроса"))
		return
	}

	var input struct{
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	err = question.ChangeWorkStatus(input.Status, input.Reason)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось изменить статус"))
		return
	}

	resp := u.Message(true, "PATH Shipment Status")
	resp["question"] = question
	u.Respond(w, resp)
}