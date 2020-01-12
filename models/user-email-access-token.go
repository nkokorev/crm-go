package models

import (
	"errors"
	"fmt"
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
	CreatedAt time.Time `json:"created_at"`
}

/*const TypeAccessEmail = {
	"verification" =
}
*/
func (uat *EmailAccessToken) Create() error {

	if uat.OwnerID <= 0 {
		return errors.New("Необходимо указать владельца токена")
	}

	if uat.DestinationEmail == "" {
		return errors.New("Необходимо указать email получателя")
	}

	uat.Token = strings.ToLower(ksuid.New().String())

	// todo debug
	if uat.OwnerID == 4 {
		uat.Token = "1ukyryxpfprxpy17i4ldlrz9kg3"
	}

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

// ### Helpers FUNC

// Верифицирует пользователя по токену и возвращает пользователя в случае успеха
func (ueat *EmailAccessToken) UserEmailVerification (user *User) error {

	// 1. проверяем, есть ли такой токен
	if err := ueat.Get(); err != nil {
		return u.Error{Message:"Указанный проверочный код не существует"}
	}

	// 2. Проверяем тип кода, который соответствует переданному токену
	if ueat.ActionType != "verification" {
		return u.Error{Message:"Не верный тип проверочного кода"}
	}

	// 3. Проверяем время жизни token
	if !time.Now().Add(-time.Hour * 24).Before(ueat.CreatedAt) {
		return u.Error{Message:"Проверочный код устарел"}
	}

	// 4. Проверяем связанность кода и пользователя по owner_id = user_id AND destination_email = user_email.
	if err := db.First(user, "id = ? AND email = ?", ueat.OwnerID, ueat.DestinationEmail).Error; err != nil || &user == nil {
		return u.Error{Message:"Проверочный код предназначен для другого пользователя"}
	}

	// 5. Если пользователь уже верифицирован, то не надо его повторно верифицировать
	if user.EmailVerifiedAt != nil {
		//return nil
		return ueat.Delete()
	}

	// 6. Если все в порядке активируем учетную запись пользователя и сохраняем данные пользователя
	timeNow := time.Now()
	user.EmailVerifiedAt = &timeNow
	if err := user.Save(); err != nil {
		return u.Error{Message:"Неудалось обновить данные верификации"}
	}

	// 7. Если все хорошо, возвращаем пользователя
	//return nil
	return ueat.Delete()
}

func (ueat *EmailAccessToken) CreatUserVerificationToken(user *User) error {

	ueat.OwnerID = user.ID
	ueat.DestinationEmail = user.Email
	ueat.ActionType = "verification"

	// 1. тут будет круто сделать проверки на наличие токена или уже подтвержденного пользователя
	if user.EmailVerifiedAt != nil {
		fmt.Println(" user.EmailVerifiedAt: ",  user.EmailVerifiedAt)
		return u.Error{Message:"Email пользователя уже подтвержден"}
	}

	// 2. Проверка существующего токена
	if !ueat.ExistEmailVerification() {
		return u.Error{Message:"Email пользователя уже подтвержден"}
	}

	return ueat.Create()

}

// Проверяет существование токена для пользователя с типом verification
func (ueat *EmailAccessToken) ExistEmailVerification () bool {
	return db.First(ueat,"owner_id = ? AND destination_email = ? AND action_type = 'verification'", ueat.OwnerID, ueat.DestinationEmail).Error == gorm.ErrRecordNotFound
}

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
	if !time.Now().Add(-time.Hour * 72).Before(ueat.CreatedAt) {
		fmt.Println("Действительно устарел!", ueat.CreatedAt , time.Now())
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
	if !time.Now().Add(-time.Hour * 72).Before(ueat.CreatedAt) {
		fmt.Println("Действительно устарел!", ueat.CreatedAt , time.Now())
		return errors.New("Код приглашения устарел")
	}

	return ueat.Delete()
	//return db.First(ueat,"token = ? AND destination_email = ? AND action_type = 'ivite-user'", ueat.Token, ueat.DestinationEmail).Error == gorm.ErrRecordNotFound
}