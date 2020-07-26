package uiApiCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func UiApiOrderCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	var input struct{
		models.Order
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

/*	order, err := account.CreateEntity(&input.Order)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания ключа"}))
		return
	}*/

	resp := u.Message(true, "POST Order Created")
	// resp["order"] = order
	u.Respond(w, resp)
}