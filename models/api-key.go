package models

import (
	"github.com/segmentio/ksuid"
	"os"
	"time"
)

type ApiKey struct {
	Token string `json:"token" gorm:"primary_key;auto_increment:false;unique_index;not null;"` // ID
	AccountID uint `json:"accountId" gorm:"index,not null"` // аккаунт-владелец ключа
	Name string `json:"name"` // имя ключа "Для сайта", "Для CRM"
	Enabled bool `json:"enabled"` // активен ли ключ
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (ApiKey) PgSqlCreate() {
	
	db.CreateTable(&ApiKey{})
	
	db.Exec("ALTER TABLE api_keys \n    ADD CONSTRAINT api_keys_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\n\n-- create unique index uix_api_keys_token_account_id ON api_keys (token,account_id);")
}

func (apiKey ApiKey) create() (*ApiKey, error)  {
	var outApiKey ApiKey

	outApiKey.AccountID = apiKey.AccountID
	outApiKey.Name = apiKey.Name
	outApiKey.Enabled = apiKey.Enabled

	outApiKey.Token = ksuid.New().String()

	if os.Getenv("APP_ENV") == "local" && apiKey.AccountID == 1 {
		outApiKey.Token = "1ukyryxpfprxpy17i4ldlrz9kg3"
	}

	err := db.Create(&outApiKey).Error

	return &outApiKey, err
}

func (apiKey ApiKey) delete () error {
	return db.Model(ApiKey{}).Where("token = ?", apiKey.Token).Delete(apiKey).Error
}

// !!! В контексте аккаунта рекомендуется использовать account.GetApiKey() с проверкой AccountID
func GetApiKey(token string) (*ApiKey, error) {
	var key ApiKey
	err := db.First(&key, "token = ?", token).Error
	return &key, err
}

func (apiKey *ApiKey) update(input ApiKey) error {
	// return db.Model(apiKey).Omit("token", "account_id", "created_at", "updated_at").Select("Name", "Enabled").Updates(&input).Error
	return db.Model(apiKey).Select("Name", "Enabled").Updates(&input).Error

}


// ######## !!!! Все что выше покрыто тестами на прямую или косвено
