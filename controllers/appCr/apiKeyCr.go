package appCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)
// CRUD functional <* love

func ApiKeyCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	input := struct {
		Name string `json:"name"`
		Enabled bool `json:"enabled"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	apiKey, err := account.ApiKeyCreate(models.ApiKey{Name: input.Name, Enabled: input.Enabled})
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания ключа"}))
		return
	}

	resp := u.Message(true, "POST Api Key Created")
	resp["apiKey"] = *apiKey
	u.Respond(w, resp)
}

func ApiKeyGetList(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
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

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	apiKeyId, err := utilsCr.GetUINTVarFromRequest(r, "apiKeyId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id"))
		return
	}

	apiKey, err := account.ApiKeyGet(apiKeyId)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке запроса", Errors: map[string]interface{}{"emailTemplates":"Не удалось получить ключ"}}))
		return
	}

	resp := u.Message(true, "GET ApiKey")
	resp["apiKey"] = apiKey
	u.Respond(w, resp)
}

func ApiKeyUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	idApiKey, err := utilsCr.GetUINTVarFromRequest(r, "id")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID шаблона"))
		return
	}

	// Get JSON-request

	var input map[string]interface{}
	                                                              
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	apiKey, err := account.ApiKeyUpdate(idApiKey, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Api Key Update")
	resp["apiKey"] = apiKey
	u.Respond(w, resp)
}

func ApiKeyDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	idApiKey, err := utilsCr.GetUINTVarFromRequest(r, "id")
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