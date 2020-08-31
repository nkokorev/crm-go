package appCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

// Ui / API there!
func EmailQueueEmailTemplateCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	var input struct{
		models.EmailQueueEmailTemplate
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}
	
	emailQueueEmailTemplate, err := account.CreateEntity(&input.EmailQueueEmailTemplate)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания серии"}))
		return
	}

	resp := u.Message(true, "POST EmailQueueEmailTemplate Created")
	resp["emailQueueEmailTemplate"] = emailQueueEmailTemplate
	u.Respond(w, resp)
}

func EmailQueueEmailTemplateGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	// ThisIs PublicID or inside
	emailQueueEmailTemplateId, err := utilsCr.GetUINTVarFromRequest(r, "emailQueueEmailTemplateId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке web site Id"))
		return
	}

	var emailQueueEmailTemplate models.EmailQueueEmailTemplate

	err = account.LoadEntity(&emailQueueEmailTemplate, emailQueueEmailTemplateId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список заказов"))
		return
	}

	resp := u.Message(true, "GET EmailQueueEmailTemplate")
	resp["email_queue_email_template"] = emailQueueEmailTemplate
	u.Respond(w, resp)
}

func EmailQueueEmailTemplateGetListPagination(w http.ResponseWriter, r *http.Request) {

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

	emailQueueId, err := utilsCr.GetUINTVarFromRequest(r, "emailQueueId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id emailQueueId"))
		return
	}

	filter := make(map[string]interface{},0)
	filter["email_queue_id"] = emailQueueId

	var total int64 = 0
	emailQueueEmailTemplates := make([]models.Entity,0)

	emailQueueEmailTemplates, total, err = account.GetPaginationListEntity(&models.EmailQueueEmailTemplate{}, offset, limit, sortBy, search, filter)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	resp := u.Message(true, "GET EmailQueueEmailTemplate Pagination List")
	resp["total"] = total
	resp["email_queue_email_templates"] = emailQueueEmailTemplates
	u.Respond(w, resp)
}

func EmailQueueEmailTemplateUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	emailQueueEmailTemplateId, err := utilsCr.GetUINTVarFromRequest(r, "emailQueueEmailTemplateId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	var emailQueueEmailTemplate models.EmailQueueEmailTemplate
	err = account.LoadEntity(&emailQueueEmailTemplate, emailQueueEmailTemplateId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	err = account.UpdateEntity(&emailQueueEmailTemplate, input)
	if err != nil {
		// fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH EmailQueueEmailTemplate Update")
	resp["email_queue_email_template"] = emailQueueEmailTemplate
	u.Respond(w, resp)
}

func EmailQueueEmailTemplateMassUpdates(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	emailQueueId, err := utilsCr.GetUINTVarFromRequest(r, "emailQueueId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id emailQueueId"))
		return
	}

	emailQueue := models.EmailQueue{}
	if err =account.LoadEntity(&emailQueue, emailQueueId); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Не найдена серия писем"}))
		return
	}

	var input struct{
		EmailQueueEmailTemplates []models.MassUpdateEmailQueueTemplate `json:"emailQueueEmailTemplates"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	if err = emailQueue.UpdateOrderEmailTemplates(input.EmailQueueEmailTemplates); err !=nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе 2"))
		return
	}

	filter := make(map[string]interface{},0)
	filter["email_queue_id"] = emailQueueId

	var total int64 = 0
	emailQueueEmailTemplates := make([]models.Entity,0)

	emailQueueEmailTemplates, total, err = account.GetPaginationListEntity(&models.EmailQueueEmailTemplate{}, 0, 100, "order", "", filter)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	resp := u.Message(true, "PATCH EmailQueueEmailTemplate MassUpdates")
	resp["email_queue_email_templates"] = emailQueueEmailTemplates
	resp["total"] = total
	u.Respond(w, resp)
}

func EmailQueueEmailTemplateDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	emailQueueEmailTemplateId, err := utilsCr.GetUINTVarFromRequest(r, "emailQueueEmailTemplateId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	var emailQueueEmailTemplate models.EmailQueueEmailTemplate
	err = account.LoadEntity(&emailQueueEmailTemplate, emailQueueEmailTemplateId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}
	if err = account.DeleteEntity(&emailQueueEmailTemplate); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении шаблона"))
		return
	}

	resp := u.Message(true, "DELETE EmailQueueEmailTemplate Successful")
	u.Respond(w, resp)
}
