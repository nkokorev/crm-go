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

/* Возвращает список шаблонов для текущего аккаунта */
func EmailTemplateGet(w http.ResponseWriter, r *http.Request) {

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

	// time.Sleep(5 * time.Second)

	resp := u.Message(true, "GET account template")
	resp["template"] = template
	u.Respond(w, resp)
}

func EmailTemplatesGetList(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	templates, err := account.EmailTemplatesList()
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке запроса", Errors: map[string]interface{}{"emailTemplates":"Не удалось получить список доменов"}}))
		return
	}

	resp := u.Message(true, "GET account templates")
	resp["emailTemplates"] = templates
	u.Respond(w, resp)
}

func EmailTemplatesDelete(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	v := &struct {
		HashId       string `json:"hashId"` // hashId for deleted accoint
	}{}
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	t, err := account.EmailTemplateGetByHashID(v.HashId)
	if err != nil || t == nil {
		u.Respond(w, u.MessageError(err, "Шаблон не найден"))
		return
	}

	err = t.Delete()
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
	resp["emailTemplates"] = templates
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
		Name       string `json:"name"` // hashId for deleted accoint
		Code       string `json:"code"` // hashId for deleted accoint
	}{}
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	tpl, err := account.CreateEmailTemplate(models.EmailTemplate{Name: v.Name, Code: string(v.Code)})
	if err != nil || tpl == nil {
		u.Respond(w, u.MessageError(err, "Ошибка при создании шаблона"))
		return
	}

	// обновляем список
	templates, err := account.GetEmailTemplates()
	if err != nil || templates == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке запроса", Errors: map[string]interface{}{"emailTemplates":"Не удалось получить список доменов"}}))
		return
	}

	resp := u.Message(true, "Email templates created")
	resp["template"] = *tpl
	resp["emailTemplates"] = templates
	u.Respond(w, resp)
}

func EmailTemplatesUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	input := &struct {
		HashId       string `json:"hashId"` // hashId for deleted accoint
		Name       string `json:"name"` //
		Code       string `json:"code"` // 
	}{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// get Email Template
	tpl, err := account.EmailTemplateGetByHashID(input.HashId)
	if err != nil || tpl == nil {
		u.Respond(w, u.MessageError(err, "Шаблон не найден"))
		return
	}

	err = account.EmailTemplateUpdate(tpl, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении шаблона"))
		return
	}

	// обновляем список
	templates, err := account.GetEmailTemplates()
	if err != nil || templates == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке запроса", Errors: map[string]interface{}{"emailTemplates":"Не удалось получить список доменов"}}))
		return
	}

	resp := u.Message(true, "Email templates created")
	resp["template"] = *tpl
	resp["emailTemplates"] = templates
	u.Respond(w, resp)
}


// ### Public function ###
func EmailTemplateShareGet(w http.ResponseWriter, r *http.Request) {

	fmt.Println("EmailTemplateShareGet")
	
	hashId, err := GetSTRVarFromRequest(r, "emailTemplateHashId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID шаблона"))
		return
	}

	template, err := (models.Account{}).EmailTemplateGetSharedByHashID(hashId)
	if err != nil || template == nil {
		u.Respond(w, u.MessageError(err, "Шаблон не найден"))
		return
	}

	w.Header().Set("Content-Type", "text/html;charset=UTF-8")
	fmt.Fprint(w, template.Code)
}