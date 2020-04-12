package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
	"time"
)

// структура для создания пользователя (чтение данных)
type InputUserData struct {
	*models.User
	NativePwd string `json:"password"` // потому что пароль из User{} не читается т.к. json -
	InviteToken string `json:"inviteToken"` // если создание через инвайт токен
}

// Вспомогательная функция чтения данных в InputUserCreate структуру
func GetDataUserRegistration(r *http.Request) (*InputUserData, error) {
	// 2. Читаем данные со входа
	var input InputUserData
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return nil, errors.New("Техническая ошибка в запросе")
	}

	// Сохраняем незашифрованный пароль в User
	input.User.Password = input.NativePwd

	return &input, nil
}

/**
* Контроллер регистрации через Ui/API
* Учитывает настройки аккаунта
 */
func UserSignUp(w http.ResponseWriter, r *http.Request) {

	if r.Context().Value("account") == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке запроса", Errors: map[string]interface{}{"account":"not load"}}))
		return
	}
	account := r.Context().Value("account").(models.Account)

	var user *models.User
	var err error

	// Проверяем разрешение на регистрацию новых пользователей через UI/API
	if !account.UiApiEnabledUserRegistration {
		u.Respond(w, u.MessageError(errors.New("Регистрация новых пользователей приостановлена")))
		return
	}

	// Читаем данные для создания пользователя
	input := struct {
		models.User
		NativePwd string `json:"password"` // потому что пароль из User{} не читается т.к. json -
		InviteToken string `json:"inviteToken"` // может присутствовать
	}{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// Проверим, все ли поля нужные поля на месте и не пустые
	if err := u.CheckNotNullFields(input.User, account.UiApiUserRegistrationRequiredFields); err != nil {
		u.Respond(w, u.MessageError(err, "Не верно заполнены поля"))
		return
	}

	// Сохраняем незашифрованный пароль в User
	input.User.Password = input.NativePwd

	// глобальная переменная для регистрации по инвайтам
	var emailToken *models.EmailAccessToken

	if account.UiApiUserRegistrationInvitationOnly {

		emailToken, err = models.GetEmailAccessToken(input.InviteToken)

		if err != nil || emailToken == nil {
			u.Respond(w, u.MessageError(u.Error{Message:"Неверный код приглашения", Errors: map[string]interface{}{"inviteToken":"Код приглашения не найден"}})) // что это?)
			return
		}

		if emailToken.Expired() {

			_ = emailToken.Delete()

			u.Respond(w, u.MessageError(u.Error{Message:"Ваш код приглашения устарел", Errors: map[string]interface{}{"inviteToken":"Используйте другой код"}})) // что это?)
			return
		}

		if input.Email != emailToken.DestinationEmail {
			u.Respond(w, u.MessageError(u.Error{Message:"Неверный код приглашения", Errors: map[string]interface{}{"inviteToken":"Код приглашения не найден"}})) // что это?)
			return
		}

		input.User.InvitedUserID = emailToken.OwnerID

		defer func() {
			if user != nil {
				emailToken.Delete()
			}
		}()
	}

	user, err = account.CreateUser(input.User)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось создать пользователя")) // что это?)
		return
	}

	// todo: тут должна быть какая-то проверка на необходимость отправки письма приглашения
	/*if ! input.EmailVerificated {
		if err := user.SendEmailVerification(); err !=nil {
			// ..
		}
	}*/

	// 2. Добавляем пользователя в аккаунт

	// todo add user to account

	// 2. создаем jwt-token для аутентификации пользователя
	//token, err := (models.JWT{UserID:user.ID}).CreateCryptoToken()
	expiresAt := time.Now().UTC().Add(time.Minute * 60).Unix()

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

	resp := u.Message(true, "POST user / User Create")
	resp["user"] = user
	resp["token"] = token
	u.Respond(w, resp)
}


// Обработка создания пользователя в рамках /{accountId}/
// Не подходит для создания пользователя в рамках UI/API т.к. не делает проверку соотвествующих переменных
func UserRegistration(w http.ResponseWriter, r *http.Request) {
	
	// 1. Получаем аккаунт, в рамках которого будет происходить создание нового пользователя
	if r.Context().Value("account") == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке запроса", Errors: map[string]interface{}{"account":"not load"}}))
		return
	}
	account := r.Context().Value("account").(*models.Account)
	if &account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке запроса", Errors: map[string]interface{}{"account":"not load"}}))
		return
	}

	// 2. Читаем данные со входа
	input, err := GetDataUserRegistration(r)
	if err != nil || input == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке запроса"}))
		return
	}

	// глобальная переменная для регистрации по инвайтам
	var emailToken *models.EmailAccessToken

	if account.UiApiUserRegistrationInvitationOnly {

		emailToken, err = models.GetEmailAccessToken(input.InviteToken)

		if err != nil || emailToken == nil {
			u.Respond(w, u.MessageError(u.Error{Message:"Неверный код приглашения", Errors: map[string]interface{}{"inviteToken":"Код приглашения не найден"}})) // что это?)
			return
		}

		if emailToken.Expired() {

			_ = emailToken.Delete()

			u.Respond(w, u.MessageError(u.Error{Message:"Ваш код приглашения устарел", Errors: map[string]interface{}{"inviteToken":"Используйте другой код"}})) // что это?)
			return
		}

		if input.Email != emailToken.DestinationEmail {
			u.Respond(w, u.MessageError(u.Error{Message:"Неверный код приглашения", Errors: map[string]interface{}{"inviteToken":"Код приглашения не найден"}})) // что это?)
			return
		}

		input.User.InvitedUserID = emailToken.OwnerID

		defer func() {
			if input.User != nil {
				emailToken.Delete()
			}
		}()
	}

	// роль = клиент
	user, err := account.CreateUser(*input.User)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при создании пользователя"))
		return
	}

	// 2. создаем jwt-token для аутентификации пользователя без запоминания дефолтного аккаунта
	token, err := account.AuthorizationUser(*user, false)
	if err != nil || token == "" {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке запроса"}))
		return
	}
	
	resp := u.Message(true, "POST user / User Create")
	resp["user"] = user
	resp["token"] = token
	u.Respond(w, resp)
}



func UserAuthByUsername(w http.ResponseWriter, r *http.Request) {
	
	// Получаем аккаунт, в который логинится пользователь
	if r.Context().Value("issuerAccount") == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Account is not valid"}))
		return
	}

	account := r.Context().Value("issuerAccount").(*models.Account)

	if account.ID < 1 {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Собираем переданные данные
	v := &struct {
		Username string `json:"username"`
		Password string `json:"password"`
		OnceLogin bool `json:"onceLogin"`
		RememberChoice bool `json:"rememberChoice"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	user, token, err := account.AuthorizationUserByUsername(v.Username, v.Password, v.OnceLogin, v.RememberChoice)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка авторизации пользователя!"))
		return
	}
	if user == nil || token == "" {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	aUsers, err := user.AccountList()
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось загрузить аккаунты")) // вообще тут нужен релогин
		return
	}

	resp := u.Message(true, "[POST] UserAuthorization - authorization was successful!")
	resp["token"] = token
	resp["user"] = user
	resp["aUsers"] = aUsers
	u.Respond(w, resp)
	
}

func UserAuthByEmail(w http.ResponseWriter, r *http.Request) {
	fmt.Println("UserAuthByEmail!")
}

func UserAuthByPhone(w http.ResponseWriter, r *http.Request) {
	fmt.Println("UserAuthByPhone!")
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

	/*token, err := user.CreateJWTToken()
	if err != nil {
		// возвращаем обычную верфикацию
		resp := u.Message(true, "Верификация прошла успешно! ...")
		u.Respond(w, resp)
		return
	}*/

	// если все хорошо, возвращаем токен и пользователя для будущей авторизации
	resp := u.Message(true, "Верификация прошла успешно!")
	//fmt.Println(token)
	resp["user"] = user
	resp["accounts"] = user.Accounts
	//resp["token"] = token
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

/*	token, err := user.CreateJWTToken()
	if err != nil {
		// возвращаем обычную верфикацию
		resp := u.Message(false, "Пароль сброшен, но не удалось создать токен авторизации")
		u.Respond(w, resp)
		return
	}*/

	// если все хорошо, возвращаем токен и пользователя для будущей авторизации
	resp := u.Message(true, "Пароль успешно сброшен")
	//resp["token"] = token // options

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

	/*token, err := user.CreateJWTToken()
	if err != nil {
		// возвращаем обычную верфикацию
		resp := u.Message(false, "Пароль сброшен, но не удалось создать токен авторизации")
		u.Respond(w, resp)
		return
	}*/

	// если все хорошо, возвращаем токен и пользователя для будущей авторизации
	resp := u.Message(true, "Новый пароль установлен")
	//resp["token"] = token // options
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
	
	if r.Context().Value("userId") == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"UserId is not valid"}))
		return
	}
	userID := r.Context().Value("userId").(uint)

	user := models.User{ID: userID}
	if err := user.LoadAccounts(); err !=nil {
		u.Respond(w, u.MessageError(err, "Неудалось найти пользователя")) // вообще тут нужен релогин
		return
	}

	// Подгружаем список из AccountUser
	aUsers, err := user.AccountList()
	if err != nil {
		u.Respond(w, u.MessageError(err, "Неудалось загрузить аккаунты")) // вообще тут нужен релогин
		return
	}

	resp := u.Message(true, "GET users/accounts")
	resp["aUsers"] = aUsers
	u.Respond(w, resp)
}

/**
* Контроллер авторизации пользователя (не аккаунта!)
 */

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
	/*token, err := user.LoginInAccount(accountID);
	if err != nil {
		u.Respond(w, u.MessageError(err, "Неудалось войти в аккаунт"))
		return
	}*/

	acc := models.Account{ID:accountID}
	if err := user.GetAccount(&acc); err != nil {
		u.Respond(w, u.MessageError(err, "Неудалось войти в аккаунт"))
		return
	}

	resp := u.Message(true, "[GET] LoginInAccount - authorization was successful!")
	//resp["token"] = token
	resp["account"] = acc
	u.Respond(w, resp)
}
