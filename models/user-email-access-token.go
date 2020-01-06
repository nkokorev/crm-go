package models

import (
	"errors"
	"github.com/segmentio/ksuid"
	"strings"
	"time"
)

type UserEmailAccessToken struct {
	Token 	string `json:"token"` // json:"token"
	Type 	string `json:"type"` // json:"verification, recover (username, password, email), join to account"
	Email 	string `json:"email"` // json:"email"
	UserID 	uint `json:"user_id" `
	CreatedAt time.Time `json:"created_at"`
}

func (uat *UserEmailAccessToken) Create() error {

	if uat.UserID <= 0 {
		return errors.New("Необходимо указать User ID")
	}

	uat.Token = strings.ToLower(ksuid.New().String())

	// todo debug
	if uat.UserID >= 4 {
		uat.Token = "1ukyryxpfprxpy17i4ldlrz9kg3"
	}

	return db.Create(uat).Error
}

// осуществляет поиск по ID
func (uat *UserEmailAccessToken) Get () error {
	return db.First(uat,"token = ?", uat.Token).Error
}

// удаляет по ID
func (uat *UserEmailAccessToken) Delete () error {
	return db.Model(ApiKey{}).Where("token = ?", uat.Token).Delete(uat).Error
}
