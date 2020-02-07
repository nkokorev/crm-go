package models

import (
	"github.com/nkokorev/crm-go/utils"
	"github.com/segmentio/ksuid"
	"os"
	"time"
)

type ApiKey struct {
	Token string `json:"token"` // ID
	AccountID uint `json:"accountId"` // кто создал его
	Name string `json:"name"` // имя ключа
	Status bool `json:"status"` // активен ли ключ
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Account Account `json:"-"`
}

// ### CRUD FUNC ###

// Создает новый ключ, генерирует token (первичный ключ)
func (key *ApiKey) create () error {

	if key.AccountID == 0 {
		return utils.Error{Message:"Ошибка при создании Api-ключа", Errors: map[string]interface{}{"apiKey":"Неудалось привязать ключ к аккаунту"}}
	}

	key.Token = ksuid.New().String()

	if os.Getenv("APP_ENV") == "local" && key.AccountID == 1 {
		key.Token = "1ukyryxpfprxpy17i4ldlrz9kg3"
	}

	return db.Create(key).Error
}

// осуществляет поиск по Token
func GetApiKey(token string) (*ApiKey, error) {
	var key ApiKey
	err := db.First(&key, "token = ?", token).Error
	return &key, err
}

func GetApiKeyPreloadAccount(token string) (key *ApiKey, err error) {
	err = db.Preload("Account").First(key, "token = ?", token).Error
	return key, err
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
func (key *ApiKey) GetWithAccount() error {
	return db.Preload("Account").First(&key, "token = ?", key.Token).Error
}