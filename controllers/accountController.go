package controllers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
	"strconv"
)

/**
* В случае успеха возвращает в теле стандартного ответа [user]
 */
func AccountCreate(w http.ResponseWriter, r *http.Request) {

	// Аккаунт, в рамках которого происходит вызов
	if r.Context().Value("issuerAccount") == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке запроса", Errors: map[string]interface{}{"account":"not load"}}))
		return
	}
	issuerAccount := r.Context().Value("issuerAccount").(models.Account)

	//time.Sleep(1 * time.Second)
	userId := r.Context().Value("userId").(uint)

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

	// Ищем пользователя, только в контексте signed-аккаунта
	user, err := issuerAccount.GetUserById(userId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Пользователь не существует"))
		return
	}

	account, err := user.CreateAccount(acc.Account)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Cant create account")) // что это?)
		return
	}

	// 1. создаем jwt-token для аутентификации пользователя
	token, err := (models.JWT{UserID:userId, AccountID:acc.ID}).CreateCryptoToken()
	if err != nil {
		u.Respond(w, u.MessageError(err, "Cant create jwt-token"))
		return
	}

	resp := u.Message(true, "POST account / Account Create")
	resp["account"] = account
	resp["token"] = token
	u.Respond(w, resp)
}

func AccountGetProfile(w http.ResponseWriter, r *http.Request) {

	// Получаем аккаунт, в который логинится пользователь
	/*if r.Context().Value("issuerAccount") == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Account is not valid"}))
		return
	}
	issuerAccount := r.Context().Value("account_id").(*models.Account)*/

	accountIdStr := mux.Vars(r)["accountId"]

	accountIdINT, err := strconv.Atoi(accountIdStr)
	if err != nil {
		u.Respond(w, u.MessageError(nil, "accountId is error"))
		return
	}
	// получаем UINT формат
	var accountId uint = uint(accountIdINT)

	account, err := models.GetAccount(accountId)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(nil, "Не удалось найти аккаунт"))
		return
	}
	
	if account.ID < 1 {
		u.Respond(w, u.MessageError(nil, "The hashID length must be 12 symbols"))
		return
	}

	resp := u.Message(true, "GET account profile")
	resp["account"] = account
	//resp["token"] = token      // нужно ли обновлять токен?
	u.Respond(w, resp)
}
