package appCr

import (
	"encoding/json"
	"fmt"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func EmailCampaignCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	var input struct{
		models.EmailCampaign
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	emailCampaign, err := account.CreateEntity(&input.EmailCampaign)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания серии"}))
		return
	}

	resp := u.Message(true, "POST EmailCampaign Created")
	resp["email_campaign"] = emailCampaign
	u.Respond(w, resp)
}

func EmailCampaignGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	// ThisIs PublicID or inside
	emailCampaignId, err := utilsCr.GetUINTVarFromRequest(r, "emailCampaignId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке web site Id"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var emailCampaign models.EmailCampaign

	// 2. Узнаем, какой id учитывается нужен
	publicOk := utilsCr.GetQueryBoolVarFromGET(r, "public_id")

	if publicOk  {
		err = account.LoadEntityByPublicId(&emailCampaign, emailCampaignId, preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
			return
		}
	} else {
		err = account.LoadEntity(&emailCampaign, emailCampaignId, preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
			return
		}
	}

	resp := u.Message(true, "GET EmailCampaign")
	resp["email_campaign"] = emailCampaign
	u.Respond(w, resp)
}

func EmailCampaignGetListPagination(w http.ResponseWriter, r *http.Request) {

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
		sortBy = "id"
	}
	if sortDesc {
		sortBy += " desc"
	}

	search, ok := utilsCr.GetQuerySTRVarFromGET(r, "search")
	if !ok {
		search = ""
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	// возвращаемые переменные
	var total int64 = 0
	emailCampaigns := make([]models.Entity,0)

	// 2. Узнаем, какой список нужен
	all := utilsCr.GetQueryBoolVarFromGET(r, "all")

	if all {
		emailCampaigns, total, err = account.GetListEntity(&models.EmailCampaign{}, sortBy,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	} else {
		emailCampaigns, total, err = account.GetPaginationListEntity(&models.EmailCampaign{}, offset, limit, sortBy, search, nil,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	}
	


	resp := u.Message(true, "GET EmailCampaign Pagination List")
	resp["total"] = total
	resp["email_campaigns"] = emailCampaigns
	u.Respond(w, resp)
}

func EmailCampaignUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	emailCampaignId, err := utilsCr.GetUINTVarFromRequest(r, "emailCampaignId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var emailCampaign models.EmailCampaign
	err = account.LoadEntity(&emailCampaign, emailCampaignId, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// Статус меняется только через отдельную функцию
	delete(input,"status")
	delete(input,"reason")

	err = account.UpdateEntity(&emailCampaign, input, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH EmailCampaign Update")
	resp["email_campaign"] = emailCampaign
	u.Respond(w, resp)
}


func EmailCampaignChangeStatus(w http.ResponseWriter, r *http.Request) {


	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	emailCampaignId, err := utilsCr.GetUINTVarFromRequest(r, "emailCampaignId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var emailCampaign models.EmailCampaign
	err = account.LoadEntity(&emailCampaign, emailCampaignId, preloads)
	if err != nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Не удалось получить кампанию"))
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

func EmailCampaignCheckDouble(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	emailCampaignId, err := utilsCr.GetUINTVarFromRequest(r, "emailCampaignId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var emailCampaign models.EmailCampaign
	err = account.LoadEntity(&emailCampaign, emailCampaignId, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить кампанию"))
		return
	}

	// Запускает кампанию прямо сейчас
	// todo: надо сделать проверку на планировку задачи
	count, err := emailCampaign.CheckDoubleFromHistory()
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка запуска кампании"))
		return
	}

	resp := u.Message(true, "GET Email Campaign Check doubles")
	resp["count"] = count
	u.Respond(w, resp)
}

func EmailCampaignDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	emailCampaignId, err := utilsCr.GetUINTVarFromRequest(r, "emailCampaignId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	var emailCampaign models.EmailCampaign
	err = account.LoadEntity(&emailCampaign, emailCampaignId, nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}
	if err = account.DeleteEntity(&emailCampaign); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении email company"))
		return
	}

	resp := u.Message(true, "DELETE EmailCampaign Successful")
	u.Respond(w, resp)
}
