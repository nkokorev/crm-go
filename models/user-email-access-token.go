package models

import (
	"errors"
	"github.com/fatih/structs"
	u "github.com/nkokorev/crm-go/utils"
	"github.com/segmentio/ksuid"
	"gorm.io/gorm"
	"strings"
	"time"
)

type EmailAccessToken struct {
	Token 			string `json:"token"` // json:"token"
	ActionType 		string `json:"action_type"` // json:"verification, recover (username, password, email), join to account"
	DestinationEmail string `json:"destination_email"` // куда отправлять email и для какого емейла был предназначен токен. Не может быть <null>, только целевые приглашения.
	OwnerId 		uint `json:"owner_id"` // userId - создатель токена (может быть self)
	NotificationCount uint `json:"notification_count"` // число успешных уведомлений
	NotificationAt 	time.Time `json:"notification_at"` // время ПОСЛЕДНЕГО уведомления
	CreatedAt 		time.Time `json:"created_at"`
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


// проверяет время жизни токена
func (eat EmailAccessToken) Expired() bool  {

	var duration time.Duration
	c := EmailTokenType

	switch eat.ActionType {
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

	return !time.Now().UTC().Add(-duration).Before(eat.CreatedAt)
}


// ### !!! все что выше покрыто тестами


// ### CRUD function
func (eat *EmailAccessToken) Create() error {

	if eat.OwnerId <= 0 {
		return errors.New("Необходимо указать владельца токена")
	}

	if eat.DestinationEmail == "" {
		return errors.New("Необходимо указать email получателя")
	}

	eat.Token = strings.ToLower(ksuid.New().String())

	// todo debug
	/*if uat.OwnerId == 4 {
		uat.Token = "1ukyryxpfprxpy17i4ldlrz9kg3"
	}*/


	return db.Create(eat).Error
}

// осуществляет поиск по token
func GetEmailAccessToken(token string) (*EmailAccessToken, error) {
	var eat EmailAccessToken
	err := db.First(&eat,"token = ?", token).Error

	return &eat, err
}
func (eat *EmailAccessToken) Get () error {
	return db.First(eat,"token = ?", eat.Token).Error
}

// удаляет по token
func (eat *EmailAccessToken) Delete () error {
	return db.Model(ApiKey{}).Where("token = ?", eat.Token).Delete(eat).Error
}

// сохраняет все поля в модели, кроме id, deleted_at
func (eat *EmailAccessToken) Update (input map[string]interface{}) error {
	//return db.Model(EmailAccessToken{}).Where("token = ?", ueat.Token).Omit("created_at").Save(ueat).Find(ueat, "token = ?", ueat.Token).Error
	return db.Model(EmailAccessToken{}).Where("token = ?", eat.Token).Omit("created_at").Updates(input).Error
}

// ### Helpers FUNC



// ### CONFIRM FUNC ### ///

// Верифицирует пользователя по токену и возвращает пользователя в случае успеха
func (eat *EmailAccessToken) UserEmailVerificationConfirm (user *User) error {

	// 1. проверяем, есть ли такой токен
	if err := eat.Get(); err != nil {
		return u.Error{Message:"Указанный проверочный код не существует"}
	}

	// 2. Проверяем тип токена (может быть любого типа верификаци)
	if eat.ActionType != EmailTokenType.USER_EMAIL_INVITE_VERIFICATION {
		return errors.New("Неверный тип кода верификации")
	}

	// 3. Проверяем время жизни token
	if eat.Expired() {
		return u.Error{Message:"Проверочный код устарел"}
	}

	// 4. Проверяем связанность кода и пользователя по owner_id = user_id AND destination_email = user_email.
	if err := db.First(user, "id = ? AND email = ?", eat.OwnerId, eat.DestinationEmail).Error; err != nil || &user == nil {
		return u.Error{Message:"Проверочный код предназначен для другого пользователя"}
	}

	// 5. Если пользователь уже верифицирован, то не надо его повторно верифицировать
	if user.EmailVerifiedAt != nil {
		// todo
		//return nil
		return eat.Delete()
	}

	// 6. Если все в порядке активируем учетную запись пользователя и сохраняем данные пользователя
	timeNow := time.Now().UTC()
	user.EmailVerifiedAt = &timeNow
	// if err := user.update(user); err != nil {
	if err := user.save(); err != nil {
		return u.Error{Message:"Не удалось обновить данные верификации"}
	}

	// 7. Если все хорошо, возвращаем пользователя
	//return nil
	return eat.Delete()
}

// Проверяет токен и сбрасывает пароль пользователю
/*func (eat *EmailAccessToken) UserPasswordResetConfirm (user *User) error {

	// 1. проверяем, есть ли такой токен
	if err := eat.Get(); err != nil {
		return u.Error{Message:"Указанный проверочный код не существует"}
	}

	// 2. Проверяем тип токена (может быть любого типа верификаци)
	if eat.ActionType != EmailTokenType.USER_EMAIL_RESET_PASSWORD {
		//return errors.New("Не верный тип кода верификации")
		return u.Error{Message:"Неверный тип токена верфикации"}
	}

	// 3. Проверяем время жизни token
	if eat.Expired() {
		return u.Error{Message:"Проверочный код устарел"}
	}

	// 4. Проверяем связанность токена и пользователя по owner_id = user_id AND destination_email = user_email.
	if err := db.First(user, "id = ? AND email = ?", eat.OwnerId, eat.DestinationEmail).Error; err != nil || &user == nil {
		return u.Error{Message:"Проверочный код предназначен для другого пользователя"}
	}

	return user.ResetPassword()
}*/

// Удаляет токен по сбросу пароля
func (EmailAccessToken) UserDeletePasswordReset(user *User) {

	// Удаляем токен, если находим
	// todo: лютая хуйня вроде как __^^__
	if err := db.Delete(&EmailAccessToken{},"(owner_id = ? OR destination_email = ?) AND action_type = ?", user.Id, user.Email, EmailTokenType.USER_EMAIL_RESET_PASSWORD).Error; err != nil {
		// log.Fatal()...
	}
}

// ### Create TOKENS ###

// Создает токен для инвайт-верификации
func (eat *EmailAccessToken) CreateInviteVerificationToken(user *User) error {

	// Надо понять, создавать новый или использовать существующий
	if err := db.First(eat,"owner_id = ? AND destination_email = ? AND action_type = ?", user.Id, user.Email, EmailTokenType.USER_EMAIL_INVITE_VERIFICATION).Error; err!=nil {
		return nil
	}

	eat.OwnerId = user.Id
	eat.DestinationEmail = *user.Email
	eat.ActionType = EmailTokenType.USER_EMAIL_INVITE_VERIFICATION
	eat.NotificationCount = 0

	return eat.Create()

}

// Создает токен для сброса пароля
func (eat *EmailAccessToken) CreateResetPasswordToken(user *User) error {

	// Надо понять, создавать новый или использовать существующий
	if err := db.First(eat,"owner_id = ? AND destination_email = ? AND action_type = ?", user.Id, user.Email, EmailTokenType.USER_EMAIL_RESET_PASSWORD).Error;err != nil {
		return nil
	}

	eat.OwnerId = user.Id
	eat.DestinationEmail = *user.Email
	eat.ActionType = EmailTokenType.USER_EMAIL_RESET_PASSWORD
	eat.NotificationCount = 0

	return eat.Create()

}


// ### Checking some state ###

// проверяет существование инвайта
func (eat *EmailAccessToken) CheckInviteToken() error {

	// 1. Пробуем найти код приглашения
	if err := db.First(eat,"token = ? AND destination_email = ? AND action_type = 'invite-user'", eat.Token, eat.DestinationEmail).Error;err != nil {

		if err == gorm.ErrRecordNotFound {
			return errors.New("Код приглашения не найден")
		} else {
			return err
		}
	}

	// 2. Проверяем время жизни token
	if !time.Now().UTC().Add(-time.Hour * 72).Before(eat.CreatedAt) {
		return errors.New("Код приглашения устарел")
	}

	return nil
}

// проверяет инвайт для новых пользователей по ключу и емейлу
// todo чем хуже функция delete ?
func (eat *EmailAccessToken) UseInviteToken(user *User) error {

	if err := db.First(eat,"token = ? AND destination_email = ? AND action_type = 'invite-user'", eat.Token, eat.DestinationEmail).Error;err != nil {

		if err == gorm.ErrRecordNotFound {
			return errors.New("Код приглашения не найден")
		} else {
			return err
		}

	}

	// 3. Проверяем время жизни token
	if !time.Now().UTC().Add(-time.Hour * 72).Before(eat.CreatedAt) {
		return errors.New("Код приглашения устарел")
	}

	return eat.Delete()
}


// ### Sending func

// Универсальная функция по отсылке уведомления на почту
func (eat *EmailAccessToken) SendMail() error {

	// Проверяем все необходимые данные
	if eat.Token == "" {
		return errors.New("Отсутствует токен для отправки")
	}

	// Проверяем время отправки
	if !eat.NotificationAt.Add(time.Minute*3).Before( time.Now().UTC()) {
		return u.Error{Message:"Подождите несколько минут, прежде чем повторить отправку"}
	}

	// Проверяем существование email'а - depricated (проверем во время отправки)
	if err := u.EmailValidation(eat.DestinationEmail); err != nil {
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
	eat.NotificationAt = time.Now().UTC()
	eat.NotificationCount++

	// if err := eat.Update(map[string]interface{}{"notificationCount":eat.NotificationCount, "notificationAt":eat.NotificationAt}); err != nil {
	if err := eat.Update(structs.Map(eat)); err != nil {
		return err
	}

	return nil
}