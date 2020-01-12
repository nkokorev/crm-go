package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
	"time"
)

/**
* В случае успеха возвращает в теле стандартного ответа [user]
 */
func UserCreate(w http.ResponseWriter, r *http.Request) {

	//time.Sleep(1 * time.Second)
	//crmSettings := r.Context().Value("crmSettings").(models.CrmSetting)

	//u.TimeTrack(time.Now())
	crmSettings, err := models.CrmSetting{}.Get()
	if err != nil {
		u.Respond(w, u.MessageError(nil, "Сервер не может обработать запрос")) // что это?)
		return
	}

	if !crmSettings.UserRegistrationAllow {
		u.Respond(w, u.MessageError(nil, "Создание новых пользователей временно приостановлено")) // что это?)
		return
	}

	user := struct {
		models.User
		NativePwd string `json:"password"`
		InviteToken string `json:"inviteToken"` //
		EmailVerificated bool `json:"emailVerificated"` //default false
	}{}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		//u.Respond(w, u.MessageError(err, "Invalid request - cant decode json request."))
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	user.Password = user.NativePwd

	// 1. Создаем пользователя
	if err := user.Create(models.UserCreateSettings{SendEmailVerification:!user.EmailVerificated}); err != nil {
		u.Respond(w, u.MessageError(err, "Cant create user")) // что это?)
		return
	}

	// 2. Добавляем пользователя в аккаунт

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

	user := &models.User{}

	AccessData := struct {
		Token string `json:"token"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&AccessData); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// пробуем пройти верификацию
	if err := user.EmailVerification(AccessData.Token); err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось пройти верификаицю email"))
		return
	}

	token, err := user.CreateJWTToken()
	if err != nil {
		// возвращаем обычную верфикацию
		resp := u.Message(true, "Верификация прошла успешно! ...")
		u.Respond(w, resp)
		return
	}

	// если все хорошо, возвращаем токен и пользователя для будущей авторизации
	resp := u.Message(true, "Верификация прошла успешно!")
	resp["user"] = user
	resp["accounts"] = user.Accounts
	resp["token"] = token
	u.Respond(w, resp)
}

func UserGetProfile(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value("user_id").(uint)

	user := models.User{ID: userID}
	if err := user.Get(); err !=nil {
		u.Respond(w, u.MessageError(err, "Неудалось найти пользователя")) // вообще тут нужен релогин
		return
	}

	resp := u.Message(true, "GET UserProfile")
	resp["user"] = user
	u.Respond(w, resp)
}

func UserGetAccounts(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value("user_id").(uint)

	user := models.User{ID: userID}
	if err := user.LoadAccounts(); err !=nil {
		u.Respond(w, u.MessageError(err, "Неудалось найти пользователя")) // вообще тут нужен релогин
		return
	}

	resp := u.Message(true, "GET users/accounts")
	resp["accounts"] = user.Accounts
	u.Respond(w, resp)
}

/**
* Контроллер авторизации пользователя (не аккаунта!)
 */
func UserAuthorization(w http.ResponseWriter, r *http.Request)  {

	time.Sleep(0 * time.Second)
	user := &models.User{}

	v := &struct {
		Username string `json:"username"`
		Password string `json:"password"`
		StaySignedIn bool `json:"stay_signed_in"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	token, err := user.AuthLogin(v.Username, v.Password, v.StaySignedIn)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// загружаем доступные аккаунты
	if len(user.Accounts) == 0 {
		fmt.Println("BSeach accounts")
		if err := user.LoadAccounts(); err !=nil {
			u.Respond(w, u.MessageError(err, "Неудалось загрузить аккаунты")) // вообще тут нужен релогин
			return
		}
	}


	resp := u.Message(true, "[POST] UserAuthorization - authorization was successful!")
	resp["token"] = token
	resp["user"] = user
	resp["accounts"] = user.Accounts
	u.Respond(w, resp)
}

/**
* Контроллер авторизации пользователя (не аккаунта!)
 */
func UserLoginInAccount(w http.ResponseWriter, r *http.Request)  {

	accountID, err := u.GetFromRequestUINT(r, "account_id")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при входе в аккаунт"))
		return
	}
	userID := r.Context().Value("user_id").(uint)

	user := models.User{ID:userID}

	// 1. Проверяем, что пользователь действителен и существует
	if err := user.Get(); err != nil {
		u.Respond(w, u.MessageError(err, "Неудалось найти пользователя")) // вообще тут нужен релогин
		return
	}


	// 2. Пробуем войти в аккаунт, возможно много ограничений (доступ, оплата и т.д.)
	token, err := user.LoginInAccount(accountID);
	if err != nil {
		u.Respond(w, u.MessageError(err, "Неудалось войти в аккаунт"))
		return
	}

	acc := models.Account{ID:accountID}
	if err := user.GetAccount(&acc); err != nil {
		u.Respond(w, u.MessageError(err, "Неудалось войти в аккаунт"))
		return
	}

	resp := u.Message(true, "[GET] LoginInAccount - authorization was successful!")
	resp["token"] = token
	resp["account"] = acc
	u.Respond(w, resp)
}
