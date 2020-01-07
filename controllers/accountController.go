package controllers

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

/**
* В случае успеха возвращает в теле стандартного ответа [user]
 */
func AccountCreate(w http.ResponseWriter, r *http.Request) {

	//time.Sleep(1 * time.Second)
	userID := r.Context().Value("user_id").(uint)

	acc := struct {
		models.Account
		//NativePwd string `json:"password"`
		//EmailVerificated bool `json:"email_verificated"` //default false
	}{}

	if err := json.NewDecoder(r.Body).Decode(&acc); err != nil {
		//u.Respond(w, u.MessageError(err, "Invalid request - cant decode json request."))
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	if err := acc.Create(); err != nil {
		u.Respond(w, u.MessageError(err, "Cant create account")) // что это?)
		return
	}

	// 1. создаем jwt-token для аутентификации пользователя
	token, err := (models.JWT{UserId:userID, AccountId:acc.ID}).CreateCryptoToken()
	if err != nil {
		u.Respond(w, u.MessageError(err, "Cant create jwt-token"))
		return
	}

	resp := u.Message(true, "POST account / Account Create")
	resp["account"] = acc.Account
	resp["token"] = token
	u.Respond(w, resp)
}

func AccountGetProfile(w http.ResponseWriter, r *http.Request) {

	accountID := r.Context().Value("account_id").(uint)

	acc := models.Account{ID: accountID}
	if err := acc.Get(); err !=nil {
		u.Respond(w, u.MessageError(err, "Неудалось найти аккаунт")) // вообще тут нужен релогин
		return
	}

	resp := u.Message(true, "GET account profile")
	resp["account"] = acc
	//resp["token"] = token
	u.Respond(w, resp)
}
