package appCr

import (
	"encoding/json"
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
	resp["emailCampaign"] = emailCampaign
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

	var emailCampaign models.EmailCampaign

	// 2. Узнаем, какой id учитывается нужен
	publicOk := utilsCr.GetQueryBoolVarFromGET(r, "publicId")

	if publicOk  {
		err = account.LoadEntityByPublicId(&emailCampaign, emailCampaignId)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
			return
		}
	} else {
		err = account.LoadEntity(&emailCampaign, emailCampaignId)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
			return
		}
	}

	resp := u.Message(true, "GET EmailCampaign")
	resp["emailCampaign"] = emailCampaign
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
		sortBy = ""
	}
	if sortDesc {
		sortBy += " desc"
	}

	search, ok := utilsCr.GetQuerySTRVarFromGET(r, "search")
	if !ok {
		search = ""
	}

	// возвращаемые переменные
	var total uint = 0
	emailCampaigns := make([]models.Entity,0)

	// 2. Узнаем, какой список нужен
	all, allOk := utilsCr.GetQuerySTRVarFromGET(r, "all")

	if all == "true" && allOk {
		emailCampaigns, total, err = account.GetListEntity(&models.EmailCampaign{}, sortBy)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	} else {
		emailCampaigns, total, err = account.GetPaginationListEntity(&models.EmailCampaign{}, offset, limit, sortBy, search, nil)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	}
	


	resp := u.Message(true, "GET EmailCampaign Pagination List")
	resp["total"] = total
	resp["emailCampaigns"] = emailCampaigns
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

	var emailCampaign models.EmailCampaign
	err = account.LoadEntity(&emailCampaign, emailCampaignId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	err = account.UpdateEntity(&emailCampaign, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH EmailCampaign Update")
	resp["emailCampaign"] = emailCampaign
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
	err = account.LoadEntity(&emailCampaign, emailCampaignId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}
	if err = account.DeleteEntity(&emailCampaign); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении магазина"))
		return
	}

	resp := u.Message(true, "DELETE EmailCampaign Successful")
	u.Respond(w, resp)
}
