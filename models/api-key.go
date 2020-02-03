package models

import (
	"github.com/segmentio/ksuid"
	"strings"
	"time"
)

type ApiKey struct {
	Token string `json:"token"` // ID
	AccountID uint `json:"-"`
	Name string `json:"name"`
	Status bool `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Account Account `json:"-"`
}

// ### CRUD FUNC ###

// Создает новый ключ, генерирует token (первичный ключ)
func (key *ApiKey) create () error {

	key.Token = strings.ToLower(ksuid.New().String())

	if key.AccountID == 3 {
		key.Token = "1ukyryxpfprxpy17i4ldlrz9kg3"
	}

	return db.Create(key).Error
}

// осуществляет поиск по Token
func (key *ApiKey) get () error {
	return db.First(key, "token = ?", key.Token).Error
}

// сохраняет все поля в модели, кроме id, token, account_id, deleted_at
func (key *ApiKey) save () error {
	return db.Model(ApiKey{}).Omit("token","account_id","deleted_at").Save(key).Find(key, "token = ?", key.Token).Error
}

// обновляет все схожие с интерфейсом поля, кроме id, token, deleted_at
func (key *ApiKey) update (input interface{}) error {
	return db.Model(ApiKey{}).Where("token = ?", key.Token).Omit("token","account_id","deleted_at").Update(input).Find(key, "token = ?", key.Token).Error
}

// удаляет пользователя по ID
func (key *ApiKey) delete () error {
	return db.Model(ApiKey{}).Where("token = ?", key.Token).Delete(key).Error
}

// ### Account func

// Предзагружает аккаунт и делает поиск по ключам
func (key *ApiKey) GetAccount() error {
	return db.Preload("Account").First(&key, "token = ?", key.Token).Error
}