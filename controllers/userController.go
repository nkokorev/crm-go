package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

/**
* В случае успеха возвращает в теле стандартного ответа [user]
 */
func UserCreate(w http.ResponseWriter, r *http.Request) {

	//time.Sleep(1 * time.Second)

	user := struct {
		models.User
		NativePwd string `json:"password"`
		EmailVerificated bool `json:"email_verificated"` //default false
	}{}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		//u.Respond(w, u.MessageError(err, "Invalid request - cant decode json request."))
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	user.Password = user.NativePwd

	if err := user.Create(!user.EmailVerificated); err != nil {
		u.Respond(w, u.MessageError(err, "Cant create user")) // что это?)
		return
	}

	// 1. Добавляем пользователя в аккаунт
	// todo add user to account

	// 2. создаем jwt-token для аутентификации пользователя
	token, err := (models.JWT{UserId:user.ID}).CreateCryptoToken()
	if err != nil {
		u.Respond(w, u.MessageError(err, "Cant create jwt-token"))
		return
	}




	resp := u.Message(true, "POST user / User Create")
	resp["user"] = user.User
	resp["token"] = token
	u.Respond(w, resp)
}

/**
* Контроллер email-верификации
 */
func UserEmailVerification(w http.ResponseWriter, r *http.Request) {

	//time.Sleep(1 * time.Second)

	AccessData := struct {
		Token string `json:"token"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&AccessData); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// пробуем пройти верификацию
	if err := (models.User{}).EmailVerified(AccessData.Token); err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось пройти верификаицю email"))
		return
	}

	resp := u.Message(true, "Верификация прошла успешно!")
	u.Respond(w, resp)
}

func UserGetProfile(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value("user_id").(uint)

	user := models.User{ID: userID}
	fmt.Println("userID: ", userID)
	if err := user.Get(); err !=nil {
		u.Respond(w, u.MessageError(err, "Неудалось найти пользователя")) // вообще тут нужен релогин
		return
	}

	resp := u.Message(true, "POST user / User Create")
	resp["user"] = user
	//resp["token"] = token
	u.Respond(w, resp)
}

func UserGetAccounts(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value("user_id").(uint)

	user := models.User{ID: userID}
	if err := user.LoadAccounts(); err !=nil {
		u.Respond(w, u.MessageError(err, "Неудалось найти пользователя")) // вообще тут нужен релогин
		return
	}

	fmt.Println(user)
	fmt.Println(user.Accounts)

	resp := u.Message(true, "GET users/accounts")
	resp["accounts"] = user.Accounts
	//resp["token"] = token
	u.Respond(w, resp)
}
