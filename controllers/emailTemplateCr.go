package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func DomainsGet(w http.ResponseWriter, r *http.Request) {
	
	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	domains, err := account.GetDomains()
	if err != nil || domains == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке запроса", Errors: map[string]interface{}{"domains":"Не удалось получить список доменов"}}))
		return
	}

	resp := u.Message(true, "GET account domains")
	resp["domains"] = domains
	u.Respond(w, resp)
}

func EmailTemplatesCreate(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	v := &struct {
		Name       string `json:"name"`
		Code       string `json:"code"`
		Public     bool `json:"public"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	template, err := account.CreateEmailTemplate(models.EmailTemplate{Name: v.Name, Code: string(v.Code)})
	if err != nil || template == nil {
		u.Respond(w, u.MessageError(err, "Ошибка при создании шаблона"))
		return
	}

	resp := u.Message(true, "Email templates created")
	resp["template"] = *template
	u.Respond(w, resp)
}

/* Возвращает список шаблонов для текущего аккаунта */
func EmailTemplateGet(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	idTemplate, err := GetUINTVarFromRequest(r, "id")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID шаблона"))
		return
	}

	template, err := account.EmailTemplateGet(idTemplate)
	if err != nil || template == nil {
		u.Respond(w, u.MessageError(err, "Шаблон не найден"))
		return
	}

	// time.Sleep(5 * time.Second)

	resp := u.Message(true, "GET email template")
	resp["template"] = template
	u.Respond(w, resp)
}

func EmailTemplatesGetList(w http.ResponseWriter, r *http.Request) {


	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	//fmt.Println(reflect.TypeOf(account).Elem()) // models.Account

	templates, err := account.EmailTemplatesList()
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке запроса", Errors: map[string]interface{}{"emailTemplates":"Не удалось получить список доменов"}}))
		return
	}

	resp := u.Message(true, "GET account templates")
	resp["emailTemplates"] = templates
	u.Respond(w, resp)
}

func EmailTemplatesUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	idTemplate, err := GetUINTVarFromRequest(r, "id")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID шаблона"))
		return
	}

	template, err := account.EmailTemplateGet(idTemplate)
	if err != nil || template == nil {
		u.Respond(w, u.MessageError(err, "Шаблон не найден"))
		return
	}

	// Get JSON-request
	input := struct {
		// HashId string `json:"hashId"` // url-vars
		Name string `json:"name"`
		Code string `json:"code"`
		Public bool `json:"public"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// err = account.EmailTemplateUpdate(tpl, input)
	err = account.EmailTemplateUpdate(template, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении шаблона"))
		return
	}

	resp := u.Message(true, "Email template update")
	resp["template"] = *template
	u.Respond(w, resp)
}

func EmailTemplatesDelete(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	idTemplate, err := GetUINTVarFromRequest(r, "id")
	if err != nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID шаблона"))
		return
	}

	template, err := account.EmailTemplateGet(idTemplate)
	if err != nil || template == nil {
		u.Respond(w, u.MessageError(err, "Шаблон не найден"))
		return
	}

	err = template.Delete()
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка удаления шаблона"))
		return
	}

	templates, err := account.GetEmailTemplates()
	if err != nil || templates == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке запроса", Errors: map[string]interface{}{"emailTemplates":"Не удалось получить список доменов"}}))
		return
	}

	resp := u.Message(true, "Email templates delete")
	resp["emailTemplates"] = templates // зачем?
	u.Respond(w, resp)
}





// -- TEST -- 
func EmailTemplateSendToUser(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	hashId, err := GetSTRVarFromRequest(r, "emailTemplateHashId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID шаблона"))
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

	ebox, err := account.GetEmailBox(input.EmailBoxId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID Email box"))
		return
	}

	// err = template.Send(*ebox,*user,input.Subject)
	err = template.SendChannel(*ebox,*user,input.Subject)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в отправке письма"))
		return
	}

	resp := u.Message(true, "GET Email Template send to user")
	u.Respond(w, resp)
}


// ### --- Public function --- ###
func EmailTemplatePreviewGetHTML(w http.ResponseWriter, r *http.Request) {

	hashId, err := GetSTRVarFromRequest(r, "emailTemplateHashId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID шаблона"))
		return
	}

	template, err := (models.Account{}).EmailTemplateGetSharedByHashID(hashId)
	if err != nil || template == nil {
		fmt.Println("Шаблон не получен..", err)
		w.Header().Set("Content-Type", "text/html;charset=UTF-8")
		w.Write([]byte(`<!DOCTYPE html><html lang="ru"><head><meta charset="UTF-8"><title>Шаблон не может быть отображен</title></head><body><h4 style="color:#5f5f5f;">Данный шаблон не может быть отображен.</h4></body></html>`))
		return
	}

	// Подготавливаем данные для шаблона
	vData, err := template.PrepareViewData(tempUser())
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

	hashId, err := GetSTRVarFromRequest(r, "emailTemplateHashId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID шаблона"))
		return
	}

	template, err := (models.Account{}).EmailTemplateGetSharedByHashID(hashId)
	if err != nil || template == nil {
		w.Header().Set("Content-Type", "text/html;charset=UTF-8")
		w.Write(errorHTMLPage("Данный шаблон не может быть отображен"))
		return
	}


	w.Header().Set("Content-Type", "text/html;charset=UTF-8")
	w.Write([]byte(template.Code))
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
