package appCr

import (
	"encoding/json"
	"fmt"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)


func EmailTemplateCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	var input struct{
		models.EmailTemplate
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// template, err := account.CreateEmailTemplate(models.EmailTemplate{Name: v.Name, Code: string(v.Code)})
	emailTemplate, err := account.CreateEntity(&input.EmailTemplate)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при создании шаблона"))
		return
	}

	resp := u.Message(true, "POST Email Templates Created")
	resp["emailTemplate"] = emailTemplate
	u.Respond(w, resp)
}

/* Возвращает список шаблонов для текущего аккаунта */
func EmailTemplateGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	emailTemplateId, err := utilsCr.GetUINTVarFromRequest(r, "emailTemplateId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID шаблона"))
		return
	}

	var emailTemplate models.EmailTemplate
	if err := account.LoadEntity(&emailTemplate,emailTemplateId); err != nil {
		u.Respond(w, u.MessageError(err, "Шаблон не найден"))
		return
	}

	// time.Sleep(5 * time.Second)

	resp := u.Message(true, "GET Email template")
	resp["emailTemplate"] = emailTemplate
	u.Respond(w, resp)
}

func EmailTemplateGetListPagination(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w, r)
	if err != nil { return }

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
		sortBy = ""
	}
	if sortDesc {
		sortBy += " desc"
	}
	search, ok := utilsCr.GetQuerySTRVarFromGET(r, "search")
	if !ok {
		search = ""
	}
	// 2. Узнаем, какой список нужен
	all, allOk := utilsCr.GetQuerySTRVarFromGET(r, "all")

	var total uint = 0
	emailTemplates := make([]models.Entity,0)

	if all == "true" && allOk {
		emailTemplates, total, err = account.GetListEntity(&models.EmailTemplate{}, sortBy)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список email-шаблонов"))
			return
		}
	} else {
		// webHooks, total, err = account.GetWebHooksPaginationList(offset, limit, search)
		emailTemplates, total, err = account.GetPaginationListEntity(&models.EmailTemplate{}, offset, limit, sortBy, search)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список email-шаблонов"))
			return
		}
	}

	resp := u.Message(true, "GET Email Template Pagination List")
	resp["total"] = total
	resp["emailTemplates"] = emailTemplates
	u.Respond(w, resp)
}

func EmailTemplateUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	emailTemplateId, err := utilsCr.GetUINTVarFromRequest(r, "emailTemplateId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID шаблона"))
		return
	}

	var emailTemplate models.EmailTemplate
	if err := account.LoadEntity(&emailTemplate, emailTemplateId); err != nil {
		u.Respond(w, u.MessageError(err, "Шаблон не найден"))
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// err = account.EmailTemplateUpdate(tpl, input)
	err = account.UpdateEntity(&emailTemplate, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении шаблона"))
		return
	}

	resp := u.Message(true, "Email Template Updated")
	resp["emailTemplate"] = emailTemplate
	u.Respond(w, resp)
}

func EmailTemplateDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	emailTemplateId, err := utilsCr.GetUINTVarFromRequest(r, "emailTemplateId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID шаблона"))
		return
	}

	var emailTemplate models.EmailTemplate
	if err := account.LoadEntity(&emailTemplate, emailTemplateId); err != nil {
		u.Respond(w, u.MessageError(err, "Шаблон не найден"))
		return
	}

	err = account.DeleteEntity(&emailTemplate)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка удаления шаблона"))
		return
	}

	resp := u.Message(true, "Email templates delete")
	u.Respond(w, resp)
}





// -- TEST -- 
func EmailTemplateSendToUser(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	hashId, ok := utilsCr.GetSTRVarFromRequest(r, "emailTemplateHashId")
	if !ok {
		u.Respond(w, u.MessageError(nil, "Ошибка в обработке ID шаблона"))
		return
	}

	template, err := account.EmailTemplateGetByHashID(hashId)
	if err != nil || template == nil {
		u.Respond(w, u.MessageError(err, "Шаблон не найден"))
		return
	}

	// 2. Get JSON-request
	input := &struct {
		UserId uint `json:"userId"`
		EmailBoxId uint `json:"emailBoxId"` // emailBoxId
		Subject string 	`json:"subject"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	user, err := account.GetUser(input.UserId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID пользователя"))
		return
	}

	var ebox models.EmailBox
	err = account.LoadEntity(&ebox, input.EmailBoxId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID Email box"))
		return
	}

	// err = template.Send(*ebox,*user,input.Subject)
	err = template.SendChannel(ebox,user.Email,input.Subject, nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в отправке письма"))
		return
	}

	resp := u.Message(true, "GET Email Template send to user")
	u.Respond(w, resp)
}


// ### --- Public function --- ###
func EmailTemplatePreviewGetHTML(w http.ResponseWriter, r *http.Request) {

	hashId, ok := utilsCr.GetSTRVarFromRequest(r, "emailTemplateHashId")
	if !ok {
		u.Respond(w, u.MessageError(nil, "Ошибка в обработке ID шаблона"))
		return
	}

	template, err := (models.Account{}).EmailTemplateGetSharedByHashID(hashId)
	if err != nil || template == nil {
		w.Header().Set("Content-Type", "text/html;charset=UTF-8")
		w.Write([]byte(`<!DOCTYPE html><html lang="ru"><head><meta charset="UTF-8"><title>Шаблон не может быть отображен</title></head><body><h4 style="color:#5f5f5f;">Данный шаблон не может быть отображен.</h4></body></html>`))
		return
	}

	// Подготавливаем данные для шаблона
	// vData, err := template.PrepareViewData(tempUser())
	vData, err := template.PrepareViewData(nil)
	if err != nil {
		w.Header().Set("Content-Type", "text/html;charset=UTF-8")
		w.Write(errorHTMLPage("Ошибка подготовки данных для отображения HTML"))
		return
	}

	html, err := template.GetHTML(vData)
	if err != nil {
		fmt.Println(err)
		w.Header().Set("Content-Type", "text/html;charset=UTF-8")
		w.Write(errorHTMLPage("Ошибка получения HTML из шаблона"))
		return
	}


	w.Header().Set("Content-Type", "text/html;charset=UTF-8")
	w.Write([]byte(html))
}

func EmailTemplatePreviewGetRawHTML(w http.ResponseWriter, r *http.Request) {

	hashId, ok := utilsCr.GetSTRVarFromRequest(r, "emailTemplateHashId")
	if !ok {
		u.Respond(w, u.MessageError(nil, "Ошибка в обработке ID шаблона"))
		return
	}

	template, err := (models.Account{}).EmailTemplateGetSharedByHashID(hashId)
	if err != nil || template == nil {
		w.Header().Set("Content-Type", "text/html;charset=UTF-8")
		w.Write(errorHTMLPage("Данный шаблон не может быть отображен"))
		return
	}


	w.Header().Set("Content-Type", "text/html;charset=UTF-8")
	w.Write([]byte(template.HTMLData))
}

func errorHTMLPage(errorText string) []byte {
	return []byte(
		fmt.Sprintf(
			`<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="UTF-8">
	<title>Шаблон не может быть отображен</title>
</head>
<body>
	<h4 style="color:#5f5f5f;">%s</h4>
</body>
</html>`,
			errorText))
}

func tempUser() models.User {
	return models.User{
		Username: "serName",
		Name: "Николай",
		Surname: "Иваньков",
		Email: "info@example.com",
		PhoneRegion: "RU",
		Phone: "+79251002030",
		Password: "kjdfhkdfsr439rrfh39f34",
	}
}
