package appCr

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"log"
	"net/http"
	"time"
)

// helper struct: for create user
type inputUserData struct {
	*models.User
	NativePwd   string `json:"password"`    // потому что пароль из User{} не читается т.к. json -
	InviteToken string `json:"inviteToken"` // если создание через инвайт токен
}

// Вспомогательная функция чтения данных в InputUserCreate структуру
func GetDataUserRegistration(r *http.Request) (*inputUserData, error) {
	// 2. Читаем данные со входа
	var input inputUserData
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return nil, errors.New("Техническая ошибка в запросе")
	}

	// Сохраняем незашифрованный пароль в User
	input.User.Password = &input.NativePwd

	return &input, nil
}

/**
* Контроллер регистрации через Ui/API
* Учитывает настройки аккаунта
 */
func UserSignUp(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	var user *models.User

	// Проверяем разрешение на регистрацию новых пользователей через UI/API
	if !account.UiApiEnabledUserRegistration {
		u.Respond(w, u.MessageError(errors.New("Регистрация новых пользователей приостановлена")))
		return
	}

	// Читаем данные для создания пользователя
	input := struct {
		models.User
		NativePwd   string `json:"password"`    // потому что пароль из User{} не читается т.к. json -
		InviteToken string `json:"inviteToken"` // может присутствовать
	}{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// Проверим, все ли поля нужные поля на месте и не пустые
	b, err := account.UiApiUserRegistrationRequiredFields.MarshalJSON()
	if err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}
	var arr []string
	err = json.Unmarshal(b, &arr)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	if err := u.CheckNotNullFields(input.User, arr); err != nil {
		u.Respond(w, u.MessageError(err, "Не верно заполнены поля"))
		return
	}

	// Сохраняем незашифрованный пароль в User
	input.User.Password = &input.NativePwd

	// глобальная переменная для регистрации по инвайтам
	var emailToken *models.EmailAccessToken

	if account.UiApiUserRegistrationInvitationOnly {

		emailToken, err = models.GetEmailAccessToken(input.InviteToken)

		if err != nil || emailToken == nil {
			u.Respond(w, u.MessageError(u.Error{Message: "Неверный код приглашения", Errors: map[string]interface{}{"inviteToken": "Код приглашения не найден"}})) // что это?)
			return
		}

		if emailToken.Expired() {

			_ = emailToken.Delete()

			u.Respond(w, u.MessageError(u.Error{Message: "Ваш код приглашения устарел", Errors: map[string]interface{}{"inviteToken": "Используйте другой код"}})) // что это?)
			return
		}

		if *input.Email != emailToken.DestinationEmail {
			u.Respond(w, u.MessageError(u.Error{Message: "Неверный код приглашения", Errors: map[string]interface{}{"inviteToken": "Код приглашения не найден"}})) // что это?)
			return
		}

		input.User.InvitedUserId = &emailToken.OwnerId

		defer func() {
			if user != nil {
				emailToken.Delete()
			}
		}()
	}

	roleClientMain, err := account.GetRoleByTag(models.RoleClient)
	if err != nil {
		log.Fatalf("Не удалось найти аккаунт: %v", err)
	}

	user, err = account.CreateUser(input.User, *roleClientMain)
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
	//token, err := (models.JWT{UserId:user.Id}).CreateCryptoToken()
	expiresAt := time.Now().UTC().Add(time.Minute * 60).Unix()

	claims := models.JWT{
		user.Id,
		account.Id,
		user.IssuerAccountId,
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

	// Аккаунт, в котором происходит авторизация: issuerAccount
	issuerAccount, err := utilsCr.GetIssuerAccount(w,r)
	if err != nil || issuerAccount == nil {
		return
	}

	// 1. Получаем аккаунт, в рамках которого будет происходить создание нового пользователя
	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	// 2. Читаем данные со входа
	input, err := GetDataUserRegistration(r)
	if err != nil || input == nil {
		u.Respond(w, u.MessageError(u.Error{Message: "Ошибка в обработке запроса"}))
		return
	}

	// глобальная переменная для регистрации по инвайтам
	var emailToken *models.EmailAccessToken

	if account.UiApiUserRegistrationInvitationOnly {

		emailToken, err = models.GetEmailAccessToken(input.InviteToken)

		if err != nil || emailToken == nil {
			u.Respond(w, u.MessageError(u.Error{Message: "Неверный код приглашения", Errors: map[string]interface{}{"inviteToken": "Код приглашения не найден"}})) // что это?)
			return
		}

		if emailToken.Expired() {

			_ = emailToken.Delete()

			u.Respond(w, u.MessageError(u.Error{Message: "Ваш код приглашения устарел", Errors: map[string]interface{}{"inviteToken": "Используйте другой код"}})) // что это?)
			return
		}

		if *input.Email != emailToken.DestinationEmail {
			u.Respond(w, u.MessageError(u.Error{Message: "Неверный код приглашения", Errors: map[string]interface{}{"inviteToken": "Код приглашения не найден"}})) // что это?)
			return
		}

		input.User.InvitedUserId = &emailToken.OwnerId

		defer func() {
			if input.User != nil {
				emailToken.Delete()
			}
		}()
	}

	// роль = клиент
	roleClientMain, err := account.GetRoleByTag(models.RoleClient)
	if err != nil {
		log.Fatalf("Не удалось найти аккаунт: %v", err)
	}
	user, err := account.CreateUser(*input.User, *roleClientMain)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при создании пользователя"))
		return
	}

	// 2. создаем jwt-token для аутентификации пользователя без запоминания дефолтного аккаунта
	token, err := account.AuthorizationUser(*user, false, issuerAccount)
	if err != nil || token == "" {
		u.Respond(w, u.MessageError(u.Error{Message: "Ошибка в обработке запроса"}))
		return
	}

	resp := u.Message(true, "POST User Create")
	resp["user"] = user
	resp["token"] = token
	u.Respond(w, resp)
}

// Авторизует пользователя по issuerAccount'e (не в самом аккаунте)
func UserAuthByUsername(w http.ResponseWriter, r *http.Request) {

	// Аккаунт, в котором происходит авторизация: issuerAccount
	issuerAccount, err := utilsCr.GetIssuerAccount(w,r)
	if err != nil || issuerAccount == nil {
		u.Respond(w, u.MessageError(err, "Ошибка во время авторизации пользователя"))
		return
	}

	// Get JSON-request
	input := &struct {
		Username       string `json:"username"`
		Password       string `json:"password"`
		OnceLogin      bool   `json:"onceLogin"`
		RememberChoice bool   `json:"rememberChoice"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	if input.Username == "" || input.Password == "" {
		u.Respond(w, u.MessageError(u.Error{Message: "Ошибка авторизации - укажите регистрационные данные"}))
		return
	}
	// Есть процесс авторизации пользователя, а есть выдача token. Лучше бы связать эти данные...
	// В каком аккаунте происходит авторизация? Где регистрируется триггер "user authorization"
	// user, token, err := issuerAccount.AuthorizationUserByUsername(v.Username, v.Password, v.OnceLogin, v.RememberChoice, issuerAccount)
	user, token, err := issuerAccount.AuthorizationUserByUsername(input.Username, input.Password, input.OnceLogin, input.RememberChoice, issuerAccount)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка авторизации пользователя!"))
		return
	}
	if user == nil || token == "" {
		u.Respond(w, u.MessageError(u.Error{Message: "Ошибка авторизации"}))
		return
	}

	user, err = issuerAccount.GetUserWithAUser(user.Id)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка авторизации пользователя!"))
		return
	}


	// Список аккаунтов
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
	if err := (&models.EmailAccessToken{Token: AccessData.Token}).UserEmailVerificationConfirm(user); err != nil {
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

	var user = models.User{Email: &AccessData.Email}

	// 1. Пробуем найти пользователя с таким email
	if err := user.GetByEmail(); err != nil {
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

	// var user = models.User{Username: jsonData.Username}

	// 1. Пробуем найти пользователя с таким email
	user, err := (models.User{}).GetByUsername(jsonData.Username);
	if err != nil {
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

// Устанавливает новый пароль


// Отправка email-кода верификации для новых пользователей.
func UserSendEmailInviteVerification(w http.ResponseWriter, r *http.Request) {

	userId := r.Context().Value("user_id").(uint)

	user := models.User{Id: userId}
	if err := user.Get(); err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти пользователя")) // вообще тут нужен релогин
		return
	}

	// отправляем данные залогиненного пользователя, если пользователь уже подтвержден
	if user.EmailVerifiedAt != nil {

		resp := u.Message(true, "Пользователь уже подтвержден")

		if err := user.LoadAccounts(); err != nil {
			u.Respond(w, resp)
			return
		}
		resp["user"] = user // обновить данные
		resp["accounts"] = user.Accounts

		u.Respond(w, resp)
		return
	}

	// Проверяем есть ли токен, если нет - создаем и отправляем
	if err := user.SendEmailVerification(); err != nil {
		/*fmt.Println(err)*/
		u.Respond(w, u.MessageError(err, "Не удалось отправить код подтверждения")) // вообще тут нужен релогин
		return
	}

	// если все хорошо, возвращаем токен и пользователя для будущей авторизации
	resp := u.Message(true, "Код подтверждения отправлен на почту")
	resp["time"] = time.Now().UTC()
	u.Respond(w, resp)
}

func UserGetProfile(w http.ResponseWriter, r *http.Request) {

	userId := r.Context().Value("user_id").(uint)

	/*userId, err := u.GetFromRequestUINT(r, "user_id")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при входе в аккаунт"))
		return
	}*/

	user := models.User{Id: userId}
	if err := user.Get(); err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти пользователя")) // вообще тут нужен релогин
		return
	}

	resp := u.Message(true, "GET UserProfile")
	resp["user"] = user
	u.Respond(w, resp)
}

func UserAccountsGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	hashId, ok := utilsCr.GetSTRVarFromRequest(r,"hashId")
	if !ok {
		u.Respond(w, u.MessageError(u.Error{Message:"Файл не найден"}))
		return
	}

	user, err := account.GetUserByHashId(hashId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти пользователя"))
	}

	// Подгружаем список из AccountUser
	aUsers, err := user.AccountList()
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось загрузить аккаунты")) // вообще тут нужен релогин
		return
	}

	resp := u.Message(true, "GET users/accounts")
	resp["aUsers"] = aUsers
	u.Respond(w, resp)
}

/**
* Контроллер авторизации пользователя (не аккаунта!)
 */
func UserLoginInAccount(w http.ResponseWriter, r *http.Request) {

	accountId, err := u.GetFromRequestUINT(r, "account_id")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при входе в аккаунт"))
		return
	}
	userId := r.Context().Value("user_id").(uint)

	user := models.User{Id: userId}

	// 1. Проверяем, что пользователь действителен и существует
	if err := user.Get(); err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти пользователя")) // вообще тут нужен релогин
		return
	}

	// 2. Пробуем войти в аккаунт, возможно много ограничений (доступ, оплата и т.д.)
	/*token, err := user.LoginInAccount(accountId);
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось войти в аккаунт"))
		return
	}*/

	acc := models.Account{Id: accountId}
	if err := user.GetAccount(&acc); err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось войти в аккаунт"))
		return
	}

	resp := u.Message(true, "[GET] LoginInAccount - authorization was successful!")
	//resp["token"] = token
	resp["account"] = acc
	u.Respond(w, resp)
}
