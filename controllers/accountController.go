package controllers

import (
	"encoding/json"
	"fmt"
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
		u.Respond(w, u.MessageError(u.Error{Message:"Не удалось получить user id"}))
		return
	}
	userID := r.Context().Value("userId").(uint)
	user, err := account.GetUserById(userID)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка обновления ключа авторизации.."}))
		return
	}

	// this is for ui/api
	token := ""

	switch r.Context().Value("issuer") {
		case "app":
			/*mAcc, err := models.GetMainAccount()
			if err != nil {
				u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
				return
			}

			token, err := mAcc.GetAuthToken(*user, true)
			if err != nil || token == "" {
				u.Respond(w, u.MessageError(u.Error{Message:"Не удалось обновить ключ авторизации"}))
				return
			}
*/
			tk, err := account.AuthorizationUser(*user, v.RememberChoice)
			if err != nil || tk == "" {
				u.Respond(w, u.MessageError(u.Error{Message:"Не удалось обновить ключ авторизации"}))
				return
			}
			token = tk
			// fmt.Printf("Получен токен: %v, \n", token)

		case "ui-api":
			tk, err := account.AuthorizationUser(*user, v.RememberChoice)
			if err != nil || tk == "" {
				u.Respond(w, u.MessageError(u.Error{Message:"Не удалось обновить ключ авторизации"}))
				return
			}
			token = tk
		default:
			fmt.Println(mux.Vars(r)["issuer"])
			u.Respond(w, u.MessageError(u.
				Error{Message:"Удостоверяющая подпись не найдена"}))
			return
	}


	/*token, err := account.AuthorizationUser(*user, v.RememberChoice)
	if err != nil || token == "" {
		u.Respond(w, u.MessageError(u.Error{Message:"Неудалось обновить ключ авторизации"}))
		return
	}*/

	// fmt.Println("Авторизация в аккаунте")
	// fmt.Println(account.Name)
	// fmt.Printf("Новый токен: %v\n", token)

	resp := u.Message(true, "GET account profile")
	resp["account"] = account
	resp["token"] = token // новый токен, который часа на 4..
	u.Respond(w, resp)
}

