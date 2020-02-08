package controllers

import (
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func CrmGetSettings(w http.ResponseWriter, r *http.Request) {

	//crmSettings, err := models.CrmSetting{}.Get()
	crmSettings, err := models.GetCrmSettings()
	if err != nil {
		u.Respond(w, u.MessageError(nil, "Сервер не может обработать запрос")) // что это?)
		return
	}



	resp := u.Message(true, "Get CRM Settings")
	resp["settings"] = crmSettings
	u.Respond(w, resp)
}
