package controllers

import (
	"encoding/json"
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

	// 1. Подгрузим файл настроек, он тут будет кстати
	/*crmSettings, err := models.CrmSetting{}.Get()
	if err != nil {
		u.Respond(w, u.MessageError(nil, "Сервер не может обработать запрос")) // что это?)
		return
	}*/

	// 2. Проверяем, что регистрация новых пользователей разрешена
	// todo загружаем данные аккаунта
	/*if !crmSettings.UserRegistrationAllow {
		u.Respond(w, u.MessageError(nil, "Создание новых пользователей временно приостановлено")) // что это?)
		return
	}*/

	user := struct {
		models.User
		NativePwd string `json:"password"`
		InviteToken string `json:"inviteToken"` //
		EmailVerificated bool `json:"emailVerificated"` //default false
	}{}


	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// Подставляем пароль для создания
	user.Password = user.NativePwd

	// 3. Создаем пользователя
	if err := user.Create(models.UserCreateOptions{SendEmailVerification:!user.EmailVerificated, InviteToken:user.InviteToken}); err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось создать пользователя")) // что это?)
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
* Контроллер проверки и применения кода email-верификации
 */
func UserEmailVerificationConfirm(w http.ResponseWriter, r *http.Request) {

	user := &models.User{}

	AccessData := struct {
		Token string `json:"token"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&AccessData); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// пробуем пройти верификацию
	if err := (&models.EmailAccessToken{Token:AccessData.Token}).UserEmailVerificationConfirm(user); err != nil {
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
	//fmt.Println(token)
	resp["user"] = user
	resp["accounts"] = user.Accounts
	resp["token"] = token
	u.Respond(w, resp)
}

func UserRecoveryUsername(w http.ResponseWriter, r *http.Request) {


	// почта пользователя, на которую надо отправить имя пользователя
	AccessData := struct {
		Email string `json:"email"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&AccessData); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	var user = models.User{Email:AccessData.Email}

	// 1. Пробуем найти пользователя с таким email
	if err := user.GetByEmail(); err !=nil {
		u.Respond(w, u.MessageError(err, "Email-адрес не найден"))
		return
	}

	// 2. Отправляем имя пользователя ему на почту
	if err := user.SendEmailRecoveryUsername(); err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось отправить сообщение на ваш email"))
		return
	}

	// если все хорошо, возвращаем токен и пользователя для будущей авторизации
	resp := u.Message(true, "Имя пользователя отправлено на ваш email")
	u.Respond(w, resp)
}

// Отправляет письмо с инструкцией по сбросу пароля, находя пользователя по username
func UserRecoveryPasswordSendMail(w http.ResponseWriter, r *http.Request) {

	jsonData := struct {
		Username string `json:"username"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&jsonData); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	var user = models.User{Username:jsonData.Username}

	// 1. Пробуем найти пользователя с таким email
	if err := user.GetByUsername(); err !=nil {
		u.Respond(w, u.MessageError(err, "Пользователь не найден"))
		return
	}

	// 2. Отправляем инструкцию по сбросу пароля на почту пользователя
	if err := user.RecoveryPasswordSendEmail(); err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось отправить сообщение на ваш email"))
		return
	}

	// если все хорошо, возвращаем токен и пользователя для будущей авторизации
	resp := u.Message(true, "Инструкция по сбросу пароля отправлена на ваш email")
	u.Respond(w, resp)
}

// сбрасывает пароль по token'у, возвращая авторизацию пользователя
func UserPasswordResetConfirm(w http.ResponseWriter, r *http.Request) {

	user := &models.User{}

	jsonData := struct {
		Token string `json:"token"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&jsonData); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// Сбрасываем пароль, если токен действителен
	if err := (&models.EmailAccessToken{Token:jsonData.Token}).UserPasswordResetConfirm(user); err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось сбросить пароль"))
		return
	}

	token, err := user.CreateJWTToken()
	if err != nil {
		// возвращаем обычную верфикацию
		resp := u.Message(false, "Пароль сброшен, но не удалось создать токен авторизации")
		u.Respond(w, resp)
		return
	}

	// если все хорошо, возвращаем токен и пользователя для будущей авторизации
	resp := u.Message(true, "Пароль успешно сброшен")
	resp["token"] = token // options

	resp["user"] = user // options for speed
	resp["accounts"] = user.Accounts // options for speed

	u.Respond(w, resp)
}

// Устанавливает новый пароль
func UserSetPassword(w http.ResponseWriter, r *http.Request)  {

	// 1. Сначала смотрим, что нам прислал пользователь
	jsonData := struct {
		PasswordNew string `json:"password_new"` // новый пароль
		PasswordOld string `json:"password_old,omitempty"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&jsonData); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// 2. Находим текущего пользователя
	userID := r.Context().Value("user_id").(uint)

	user := models.User{ID: userID}
	if err := user.Get(); err !=nil {
		u.Respond(w, u.MessageError(err, "Неудалось найти пользователя")) // вообще тут нужен релогин
		return
	}

	// 3. Устанавливаем новый пароль
	if err := user.SetPassword(jsonData.PasswordNew, jsonData.PasswordOld); err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось установить пароль"))
		return
	}

	token, err := user.CreateJWTToken()
	if err != nil {
		// возвращаем обычную верфикацию
		resp := u.Message(false, "Пароль сброшен, но не удалось создать токен авторизации")
		u.Respond(w, resp)
		return
	}

	// если все хорошо, возвращаем токен и пользователя для будущей авторизации
	resp := u.Message(true, "Новый пароль установлен")
	resp["token"] = token // options
	resp["user"] = user // options for speed

	u.Respond(w, resp)
}

// Отправка email-кода верификации для новых пользователей.
func UserSendEmailInviteVerification(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value("user_id").(uint)

	user := models.User{ID: userID}
	if err := user.Get(); err !=nil {
		u.Respond(w, u.MessageError(err, "Неудалось найти пользователя")) // вообще тут нужен релогин
		return
	}

	// отправляем данные залогиненного пользователя, если пользователь уже подтвержден
	if user.EmailVerifiedAt != nil {

		resp := u.Message(true, "Пользователь уже подтвержден")

		if err := user.LoadAccounts(); err  != nil {
			u.Respond(w, resp)
			return
		}
		resp["user"] = user // обновить данные
		resp["accounts"] = user.Accounts

		u.Respond(w, resp)
		return
	}

	// Проверяем есть ли токен, если нет - создаем и отправляем
	if err := user.SendEmailVerification(); err !=nil {
		/*fmt.Println(err)*/
		u.Respond(w, u.MessageError(err, "Неудалось отправить код подтверждения")) // вообще тут нужен релогин
		return
	}

	// если все хорошо, возвращаем токен и пользователя для будущей авторизации
	resp := u.Message(true, "Код подтверждения отправлен на почту")
	resp["time"] = time.Now().UTC()
	u.Respond(w, resp)
}



func UserGetProfile(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value("user_id").(uint)

	/*userID, err := u.GetFromRequestUINT(r, "user_id")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при входе в аккаунт"))
		return
	}*/

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
		//StaySignedIn bool `json:"staySignedIn"`
		OnceLogin bool `json:"onceLogin"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	token, err := user.AuthLogin(v.Username, v.Password, v.OnceLogin)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// загружаем доступные аккаунты
	if len(user.Accounts) == 0 {
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
* NEW!!! Контроллер авторизации по токену. В зависимости от типа токена, может происходит:
* - обычная одноразовая авторизация
* - одноразовая авторизация со сбрасыванием пароля
 */
func UserTokenAuthorization(w http.ResponseWriter, r *http.Request)  {

	// todo ...
	user := &models.User{}

	v := &struct {
		Username string `json:"username"`
		Password string `json:"password"`
		//StaySignedIn bool `json:"staySignedIn"`
		OnceLogin bool `json:"onceLogin"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	token, err := user.AuthLogin(v.Username, v.Password, v.OnceLogin)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// загружаем доступные аккаунты
	if len(user.Accounts) == 0 {
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
