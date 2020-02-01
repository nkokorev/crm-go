package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	u "github.com/nkokorev/crm-go/utils"
	"github.com/segmentio/ksuid"
	"strings"
	"time"
)

type EmailAccessToken struct {
	Token 	string `json:"token"` // json:"token"
	ActionType 	string `json:"action_type"` // json:"verification, recover (username, password, email), join to account"
	DestinationEmail string `json:"destination_email"` // куда отправлять email и для какого емейла был предназначен токен. Не может быть <null>, только целевые приглашения.
	OwnerID 	uint `json:"owner_id" ` // userID - создатель токена (может быть self)
	NotificationCount uint `json:"notification_count"` // число успешных уведомлений
	NotificationAt time.Time `json:"notification_at"` // время уведомления
	CreatedAt time.Time `json:"created_at"`
}

var EmailTokenType = struct {
	//USER_EMAIL_VERIFICATION string
	USER_EMAIL_INVITE_VERIFICATION string
	USER_EMAIL_PERSONAL_INVITE string
	USER_EMAIL_RESET_PASSWORD string
}{
	USER_EMAIL_INVITE_VERIFICATION: "invite-verification",
	USER_EMAIL_PERSONAL_INVITE: "personal-invite",
	USER_EMAIL_RESET_PASSWORD: "reset-password",
}




// ### CRUD function
func (uat *EmailAccessToken) Create() error {

	if uat.OwnerID <= 0 {
		return errors.New("Необходимо указать владельца токена")
	}

	if uat.DestinationEmail == "" {
		return errors.New("Необходимо указать email получателя")
	}

	uat.Token = strings.ToLower(ksuid.New().String())

	// todo debug
	/*if uat.OwnerID == 4 {
		uat.Token = "1ukyryxpfprxpy17i4ldlrz9kg3"
	}*/


	return db.Create(uat).Error
}

// осуществляет поиск по token
func (uat *EmailAccessToken) Get () error {
	return db.First(uat,"token = ?", uat.Token).Error
}

// удаляет по token
func (uat *EmailAccessToken) Delete () error {
	return db.Model(ApiKey{}).Where("token = ?", uat.Token).Delete(uat).Error
}

// сохраняет все поля в модели, кроме id, deleted_at
func (ueat *EmailAccessToken) Update (input interface{}) error {
	//return db.Model(EmailAccessToken{}).Where("token = ?", ueat.Token).Omit("created_at").Save(ueat).Find(ueat, "token = ?", ueat.Token).Error
	return db.Model(EmailAccessToken{}).Where("token = ?", ueat.Token).Omit("created_at").Update(input).Error
}

// ### Helpers FUNC

// проверяет время жизни токена
func (ueat EmailAccessToken) isExpired() bool  {

	var duration time.Duration
	c := EmailTokenType

	switch ueat.ActionType {
		case c.USER_EMAIL_INVITE_VERIFICATION:
			duration = time.Hour * 48
			break
		case c.USER_EMAIL_PERSONAL_INVITE:
			duration = time.Hour * 48
			break
		case c.USER_EMAIL_RESET_PASSWORD:
			duration = time.Hour * 24
			break
		default:
			duration = time.Hour * 3
			break
	}

	return !time.Now().UTC().Add(-duration).Before(ueat.NotificationAt)

}

// ### CONFIRM FUNC ### ///

// Верифицирует пользователя по токену и возвращает пользователя в случае успеха
func (ueat *EmailAccessToken) UserEmailVerificationConfirm (user *User) error {

	// 1. проверяем, есть ли такой токен
	if err := ueat.Get(); err != nil {
		return u.Error{Message:"Указанный проверочный код не существует"}
	}

	// 2. Проверяем тип токена (может быть любого типа верификаци)
	if ueat.ActionType != EmailTokenType.USER_EMAIL_INVITE_VERIFICATION {
		return errors.New("Не верный тип кода верификации")
	}

	// 3. Проверяем время жизни token
	if ueat.isExpired() {
		return u.Error{Message:"Проверочный код устарел"}
	}

	// 4. Проверяем связанность кода и пользователя по owner_id = user_id AND destination_email = user_email.
	if err := db.First(user, "id = ? AND email = ?", ueat.OwnerID, ueat.DestinationEmail).Error; err != nil || &user == nil {
		return u.Error{Message:"Проверочный код предназначен для другого пользователя"}
	}

	// 5. Если пользователь уже верифицирован, то не надо его повторно верифицировать
	if user.EmailVerifiedAt != nil {
		// todo
		//return nil
		return ueat.Delete()
	}

	// 6. Если все в порядке активируем учетную запись пользователя и сохраняем данные пользователя
	timeNow := time.Now().UTC()
	user.EmailVerifiedAt = &timeNow
	if err := user.Save(); err != nil {
		return u.Error{Message:"Неудалось обновить данные верификации"}
	}

	// 7. Если все хорошо, возвращаем пользователя
	//return nil
	return ueat.Delete()
}

// Проверяет токен и сбрасывает пароль пользователю
func (ueat *EmailAccessToken) UserPasswordResetConfirm (user *User) error {

	// 1. проверяем, есть ли такой токен
	if err := ueat.Get(); err != nil {
		return u.Error{Message:"Указанный проверочный код не существует"}
	}

	// 2. Проверяем тип токена (может быть любого типа верификаци)
	if ueat.ActionType != EmailTokenType.USER_EMAIL_RESET_PASSWORD {
		//return errors.New("Не верный тип кода верификации")
		return u.Error{Message:"Не верный тип токена верфикации"}
	}

	// 3. Проверяем время жизни token
	if ueat.isExpired() {
		return u.Error{Message:"Проверочный код устарел"}
	}

	// 4. Проверяем связанность токена и пользователя по owner_id = user_id AND destination_email = user_email.
	if err := db.First(user, "id = ? AND email = ?", ueat.OwnerID, ueat.DestinationEmail).Error; err != nil || &user == nil {
		return u.Error{Message:"Проверочный код предназначен для другого пользователя"}
	}

	return user.ResetPassword()
}

// Удаляет токен по сбросу пароля
func (EmailAccessToken) UserDeletePasswordReset(user *User) {

	// Удаляем токен, если находим
	if !db.Delete(EmailAccessToken{},"(owner_id = ? OR destination_email = ?) AND action_type = ?", user.ID, user.Email, EmailTokenType.USER_EMAIL_RESET_PASSWORD).RecordNotFound() {
		// log.Fatal()...
	}
}

// ### Create TOKENS ###

// Создает токен для инвайт-верификации
func (ueat *EmailAccessToken) CreateInviteVerificationToken(user *User) error {

	// Надо понять, создавать новый или использовать существующий
	if !db.First(ueat,"owner_id = ? AND destination_email = ? AND action_type = ?", user.ID, user.Email, EmailTokenType.USER_EMAIL_INVITE_VERIFICATION).RecordNotFound() {
		return nil
	}

	ueat.OwnerID = user.ID
	ueat.DestinationEmail = user.Email
	ueat.ActionType = EmailTokenType.USER_EMAIL_INVITE_VERIFICATION
	ueat.NotificationCount = 0

	return ueat.Create()

}

// Создает токен для сброса пароля
func (ueat *EmailAccessToken) CreateResetPasswordToken(user *User) error {

	// Надо понять, создавать новый или использовать существующий
	if !db.First(ueat,"owner_id = ? AND destination_email = ? AND action_type = ?", user.ID, user.Email, EmailTokenType.USER_EMAIL_RESET_PASSWORD).RecordNotFound() {
		return nil
	}

	ueat.OwnerID = user.ID
	ueat.DestinationEmail = user.Email
	ueat.ActionType = EmailTokenType.USER_EMAIL_RESET_PASSWORD
	ueat.NotificationCount = 0

	return ueat.Create()

}


// ### Checking some state ###

// проверяет существование инвайта
func (ueat *EmailAccessToken) CheckInviteToken() error {

	// 1. Пробуем найти код приглашения
	if err := db.First(ueat,"token = ? AND destination_email = ? AND action_type = 'invite-user'", ueat.Token, ueat.DestinationEmail).Error;err != nil {

		if err == gorm.ErrRecordNotFound {
			return errors.New("Код приглашения не найден")
		} else {
			return err
		}
	}

	// 2. Проверяем время жизни token
	if !time.Now().UTC().Add(-time.Hour * 72).Before(ueat.CreatedAt) {
		return errors.New("Код приглашения устарел")
	}

	return nil
}

// проверяет инвайт для новых пользователей по ключу и емейлу
func (ueat *EmailAccessToken) UseInviteToken(user *User) error {

	if err := db.First(ueat,"token = ? AND destination_email = ? AND action_type = 'invite-user'", ueat.Token, ueat.DestinationEmail).Error;err != nil {

		if err == gorm.ErrRecordNotFound {
			return errors.New("Код приглашения не найден")
		} else {
			return err
		}

	}

	// 3. Проверяем время жизни token
	if !time.Now().UTC().Add(-time.Hour * 72).Before(ueat.CreatedAt) {
		return errors.New("Код приглашения устарел")
	}

	return ueat.Delete()
}


// ### Sending func

// Универсальная функция по отсылке уведомления на почту
func (ueat *EmailAccessToken) SendMail() error {

	// Проверяем все необходимые данные
	if ueat.Token == "" {
		return errors.New("Отсутствует токен для отправки")
	}

	// Проверяем время отправки
	if !ueat.NotificationAt.Add(time.Minute*3).Before( time.Now().UTC()) {
		return u.Error{Message:"Подождите несколько минут, прежде чем повторить отправку"}
	}

	// Проверяем существование email'а - depricated (проверем во время отправки)
	if err := u.ValidateEmailDeepHost(ueat.DestinationEmail); err != nil {
		return err
	}


	// Отправляем транзакционное сообщение
	// В зависимости от типа токена отправляется разный URL для верификации чего бы ни было.
	// обычная верификация: /login/email-verification?t=<token>
	// инвайт верификация: /login/sign-up/email-verification?t=<token>
	// todo sending mail to email & type...

	// using EmailNotification..
	// 1.


	// Обновляем время
	ueat.NotificationAt = time.Now().UTC()
	ueat.NotificationCount++

	if err := ueat.Update(ueat); err != nil {
		return err
	}

	return nil
}