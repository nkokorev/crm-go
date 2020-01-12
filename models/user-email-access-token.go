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
	DestinationEmail string `json:"email"` // куда отправлять email и для какого емейла был предназначен токен. Не может быть <null>, только целевые приглашения.
	OwnerID 	uint `json:"owner_id" ` // userID - создатель токена (может быть self)
	CreatedAt time.Time `json:"created_at"`
}

func (uat *UserEmailAccessToken) Create() error {

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
func (uat *UserEmailAccessToken) Get () error {
	return db.First(uat,"token = ?", uat.Token).Error
}

// удаляет по token
func (uat *UserEmailAccessToken) Delete () error {
	return db.Model(ApiKey{}).Where("token = ?", uat.Token).Delete(uat).Error
}
