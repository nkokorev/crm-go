package controllers

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func UserCreate(w http.ResponseWriter, r *http.Request) {

	user := struct {
		models.User
		NativePwd string `json:"password"`
		SendMail bool `json:"send_mail"`
	}{}


	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		//u.Respond(w, u.MessageError(err, "Invalid request - cant decode json request."))
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	user.Password = user.NativePwd

	if err := user.Create(); err != nil {
		u.Respond(w, u.MessageError(err, "Cant create user"))
		return
	}

	resp := u.Message(true, "POST user / User Create")
	resp["user"] = user.User
	u.Respond(w, resp)
}
