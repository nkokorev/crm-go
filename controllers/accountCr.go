package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
	"time"
)

// ############ CRUD Functional ############
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
	user, err := issuerAccount.GetUser(userId)
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

func AccountUpdate(w http.ResponseWriter, r *http.Request)  {

	// 1. Получаем рабочий аккаунт, сверяем его с переданным {hashId}.
	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	hashId, err := GetSTRVarFromRequest(r, "hashId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID шаблона"))
		return
	}

	// 2. Проверяем hashID изменяемого аккаунта, совпадает ли он с авторизацией.
	// Из одного аккаунта (НЕ RatusCRM) нельзя изменить другой 
	if !account.IsMainAccount() && account.HashID != hashId {
		u.Respond(w, u.MessageError(err, "Ошибка доступа к аккаунту"))
		return
	}

	input := map[string]interface{}{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	err = account.Update(input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении аккаунта"))
		return
	}

	resp := u.Message(true, "UPDATE Account successful")
	resp["account"] = *account
	u.Respond(w, resp)
}
// ############ END OF CRUD Functional ############

// ######## USERS ##########
func AccountUserList(w http.ResponseWriter, r *http.Request) {
	// 1. Получаем рабочий аккаунт (автома. сверка с {hashId}.)
	account, err := GetWorkAccountCheckHashId(w,r)
	if err != nil || account == nil {
		return
	}

	// 2. Узнаем, какой список нужен
	limit, ok := GetQueryINTVarFromGET(r, "limit")
	if !ok {
		//u.Respond(w, u.Message(false, "Error parse limit var"))
		//return
	}
	offset, ok := GetQueryINTVarFromGET(r, "offset")
	if !ok {
		//u.Respond(w, u.Message(false, "Error parse offset var"))
		//return
	}

	typeUsers := r.URL.Query().Get("types")

	fmt.Printf("Limit %d\n", limit)
	fmt.Printf("Offset %d\n", offset)
	fmt.Printf("TypeUsers %d\n", typeUsers)



	resp := u.Message(true, "GET Account User List")
	u.Respond(w, resp)
}

// Возвращает профиль аккаунта, указанного в переменной .../{hashId}/...
func AccountAuthUser(w http.ResponseWriter, r *http.Request) {

	// Аккаунт, в котором происходит авторизация: issuerAccount
	issuerAccount, err := GetIssuerAccount(w,r)
	if err != nil || issuerAccount == nil {
		u.Respond(w, u.MessageError(nil, "Не удалось найти аккаунт"))
		return
	}

	hashId, err := GetSTRVarFromRequest(r, "hashId")
	if err != nil || hashId == "" {
		u.Respond(w, u.MessageError(nil, "Не удалось получить hashId аккаунта"))
		return
	}

	account, err := models.GetAccountByHash(hashId)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(nil, "Не удалось найти аккаунт"))
		return
	}

	/*account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(nil, "Не удалось найти аккаунт"))
		return
	}*/

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
	user, err := account.GetUser(userID)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка обновления ключа авторизации.."}))
		return
	}

	token, err := account.AuthorizationUser(*user, v.RememberChoice, issuerAccount)
	if err != nil || token == "" {
		u.Respond(w, u.MessageError(u.Error{Message:"Не удалось обновить ключ авторизации"}))
		return
	}

	resp := u.Message(true, "POST Account Auth User")
	resp["account"] = account
	resp["token"] = token // новый токен, который часа на 4..
	u.Respond(w, resp)
}

func AccountGetProfile(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(nil, "Не удалось найти аккаунт"))
		return
	}

	resp := u.Message(true, "GET account profile")
	resp["account"] = account
	u.Respond(w, resp)
}


