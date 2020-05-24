package controllers

import (
	"encoding/json"
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

	apiKeys, err := account.ApiKeysList()
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке запроса"}))
		return
	}

	resp := u.Message(true, "GET list Api keys of account")
	resp["apiKeys"] = apiKeys
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

	idApiKey, err := GetUINTVarFromRequest(r, "id")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID шаблона"))
		return
	}

	// Get JSON-request
	input := struct {
		Name string `json:"name"`
		Enabled bool `json:"enabled"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	apiKey, err := account.ApiKeyUpdate(idApiKey, &input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Api Key Update")
	resp["apiKey"] = apiKey
	u.Respond(w, resp)
}

func ApiKeyGetDelete(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	idApiKey, err := GetUINTVarFromRequest(r, "id")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID шаблона"))
		return
	}

	err = account.ApiKeyDelete(idApiKey)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении Api-ключа"))
		return
	}

	resp := u.Message(true, "DELETE Api Key Update")
	u.Respond(w, resp)
}

// END CRUD functional