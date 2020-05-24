package controllers

import (
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)
// CRUD functional <* love

func ApiKeyGetCreate(w http.ResponseWriter, r *http.Request) {

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

func ApiKeyGetList(w http.ResponseWriter, r *http.Request) {

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

func ApiKeyGet(w http.ResponseWriter, r *http.Request) {

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

func ApiKeyGetUpdate(w http.ResponseWriter, r *http.Request) {

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

func ApiKeyGetDelete(w http.ResponseWriter, r *http.Request) {

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

// END CRUD functional