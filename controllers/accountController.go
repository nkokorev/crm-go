package controllers

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
	"strconv"
	"time"
)

func AccountCreate(w http.ResponseWriter, r *http.Request) {

	// Аккаунт, от имени которого выступает пользователь
	if r.Context().Value("issuerAccount") == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке запроса", Errors: map[string]interface{}{"account":"not load"}}))
		return
	}
	issuerAccount := r.Context().Value("issuerAccount").(*models.Account)
	
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
	//token, err := (models.JWT{UserID:userId, AccountID:acc.ID}).CreateCryptoToken()
	expiresAt := time.Now().UTC().Add(time.Minute * 20).Unix()

	claims := models.JWT{
		user.ID,
		account.ID,
		user.IssuerAccountID,
		jwt.StandardClaims{
			ExpiresAt: expiresAt,
			Issuer:    "AppServer",
		},
	}

	token, err := account.GetAuthTokenWithClaims(claims)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Cant create jwt-token"))
		return
	}

	resp := u.Message(true, "POST account / Account Create")
	resp["account"] = account
	resp["token"] = token
	u.Respond(w, resp)
}

// Возвращает профиль аккаунта, указанного в переменной .../{accountId}/...
func AccountAuthUser(w http.ResponseWriter, r *http.Request) {

	accIdSTR := mux.Vars(r)["accountId"]

	accountIdINT, err := strconv.Atoi(accIdSTR)
	if err != nil {
		u.Respond(w, u.MessageError(nil, "accountId is error"))
		return
	}
	// форматируем в UINT
	var accountId uint = uint(accountIdINT)

	account, err := models.GetAccount(accountId)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(nil, "Не удалось найти аккаунт"))
		return
	}

	// Чекаем, что ID настоящего аккаунта
	if account.ID < 1 {
		u.Respond(w, u.MessageError(nil, "The hashID length must be 12 symbols"))
		return
	}

	// Читаем косвенные данные логина в аккаунте
	v := &struct {
		RememberChoice bool `json:"rememberChoice"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// Получаем token для пользователя
	if r.Context().Value("userId") == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"UserId is not valid"}))
		return
	}
	userID := r.Context().Value("userId").(uint)
	user, err := account.GetUserById(userID)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка обновления ключа авторизации.."}))
		return
	}

	token, err := account.AuthorizationUser(*user, v.RememberChoice)
	if err != nil || token == "" {
		u.Respond(w, u.MessageError(u.Error{Message:"Неудалось обновить ключ авторизации"}))
		return
	}

	resp := u.Message(true, "GET account profile")
	resp["account"] = account
	resp["token"] = token // новыйт токен, который часа на 4..
	u.Respond(w, resp)
}
