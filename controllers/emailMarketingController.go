package controllers

import (
	"encoding/json"
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
func EmailTemplatesGet(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	templates, err := account.GetEmailTemplates()
	if err != nil || templates == nil {
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

	t, err := account.GetEmailTemplateByHashID(v.HashId)
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