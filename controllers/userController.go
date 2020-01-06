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
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Cant create user")) // что это?)
		return
	}

	// 1. создаем token для email-verification
	// 2. создаем jwt-token для аутентификации пользователя


	resp := u.Message(true, "POST user / User Create")
	resp["user"] = user.User
	resp["token"] = "fdshfsdfshjKdskfdKDFjocvmidsifjiIfjhosfdsd"
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
		u.Respond(w, u.MessageError(err, "Верификация email провалена"))
		return
	}

	// создаем короткий token для пользователя (?)

	resp := u.Message(true, "Верификация прошла успешно!")
	//resp["user"] = user.User
	//resp["token"] = v // что этО?)
	u.Respond(w, resp)
}
