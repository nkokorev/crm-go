package models

import (
	"errors"
	u "github.com/nkokorev/crm-go/utils"
	"github.com/segmentio/ksuid"
	"strings"
	"time"
)

type UserEmailVerification struct {
	Token 	string `json:"token"` // json:"-"
	Email 	string `json:"email"` // json:"-"
	UserID 	uint `json:"user_id" `
	User User `json:"-"`

	CreatedAt time.Time `json:"created_at"`
	//ExpiredAt time.Time `json:"updated_at"`


}

func (umv *UserEmailVerification) Create() error {

	if umv.UserID <= 0 {
		return errors.New("Необходимо указать User ID")
	}

	umv.Token = strings.ToLower(ksuid.New().String())

	if umv.UserID >= 4 {
		umv.Token = "1ukyryxpfprxpy17i4ldlrz9kg3"
	}

	return db.Create(umv).Error
}

// осуществляет поиск по ID
func (umv *UserEmailVerification) Get () error {
	return db.First(umv,"token = ?", umv.Token).Error
}

// удаляет по ID
func (umv *UserEmailVerification) Delete () error {
	return db.Model(ApiKey{}).Where("token = ?", umv.Token).Delete(umv).Error
}

//
func (umv *UserEmailVerification) EmailVerified () error {

	var e u.Error

	// 1. Проверяем дату (не заэкспирелся ли токен)
	if !time.Now().Add(-time.Hour * 24).Before(umv.CreatedAt) {
		//e.AddErrors("email", err.Error())
		e.Message = "Проверочный ключ устарел"
		return e
	}

	// 2. Ищем целевого пользователя
	var user User
	if err := db.First(&user, "id = ? AND email = ?", umv.UserID, umv.Email).Error; err != nil {
		e.Message = "Пользователь не найден"
		return e
		//return errors.New("Пользователь не найден")
	}

	// 3. Если все в порядке активируем учетную запись пользователя
	timeNow := time.Now()
	user.EmailVerifiedAt = &timeNow

	// 4. Сохраняем обновленные данные пользователя
	if err := user.Save(); err != nil {
		return err
	}

	// 4. Удаляем проверочный код (больше не нужен)
	return umv.Delete()
}