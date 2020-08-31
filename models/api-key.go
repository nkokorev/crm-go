package models

import (
	"errors"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"log"
	"strings"
	"time"
)

type ApiKey struct {
	Id     	uint   `json:"id" gorm:"primaryKey"`
	Token 	string `json:"token" gorm:"unique_index;not null;"` // Id
	AccountId uint `json:"account_id" gorm:"index;not null"` // аккаунт-владелец ключа

	Name 	string `json:"name" gorm:"type:varchar(255);default:'New api key';"` // имя ключа "Для сайта", "Для CRM"
	Enabled bool `json:"enabled" gorm:"type:bool;default:true"` // активен ли ключ

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ApiKey) PgSqlCreate() {
	
	db.Migrator().CreateTable(&ApiKey{})
	// db.Model(&ApiKey{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE api_keys ADD CONSTRAINT api_keys_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}

func (apiKey *ApiKey) BeforeCreate(tx *gorm.DB) error {

	apiKey.Id = 0

	// 5c0511936507b48cbbf245cd080b9d2f - MailChimp
	// ekll44e6s2ro8g0hc0j5yx560e2a6zku - RatusCRM
	// apiKey.Token = ksuid.New().String()
	apiKey.Token = strings.ToLower(utils.RandStringBytesMaskImprSrcUnsafe(32, true))

	apiKey.CreatedAt = time.Now().UTC()
	return nil
}

func (ApiKey) create(input ApiKey) (*ApiKey, error)  {
	var apiKey = input
	err := db.Create(&apiKey).Error
	return &apiKey, err
}

func (ApiKey) get(id uint) (*ApiKey, error) {

	apiKey := ApiKey{}

	err := db.First(&apiKey, id).Error
	if err != nil {
		return nil, err
	}
	
	return &apiKey, nil
}

func (ApiKey) getByToken(token string) (*ApiKey, error) {

	apiKey := ApiKey{}

	err := db.First(&apiKey, "token = ?", token).Error
	if err != nil {
		return nil, err
	}

	return &apiKey, nil
}

func GetApiKeyByToken(token string) (*ApiKey, error) {
	return ApiKey{}.getByToken(token)
}

func (ApiKey) getList(accountId uint) ([]ApiKey, error) {

	apiKeys := make([]ApiKey,0)

	err := db.Find(&apiKeys, "account_id = ?", accountId).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return apiKeys, nil
}

func (apiKey *ApiKey) delete () error {
	return db.Model(ApiKey{}).Where("id = ?", apiKey.Id).Delete(apiKey).Error
}

func (apiKey *ApiKey) update(input map[string]interface{}) error {
	// return db.Model(apiKey).Omit("token", "account_id", "created_at", "updated_at").Select("Name", "Enabled").Updates(&input).Error
	return db.Model(apiKey).Select("Name", "Enabled").Updates(input).Error

}

// ######## !!!! Все что выше покрыто тестами на прямую или косвено

// ########### ACCOUNT FUNCTIONAL ###########

func (account Account) ApiKeyCreate(input ApiKey) (*ApiKey, error) {
	input.AccountId = account.Id
	return ApiKey{}.create(input)
}

func (account Account) ApiKeyGet(id uint) (*ApiKey, error) {
	apiKey, err := ApiKey{}.get(id)
	if err != nil {
		return nil, err
	}

	if apiKey.AccountId != account.Id {
		return nil, errors.New("ApiKey не принадлежит аккаунту")
	}

	return apiKey, nil
}

func (account Account) ApiKeyGetByToken(token string) (*ApiKey, error) {
	apiKey, err := GetApiKeyByToken(token)
	if err != nil {
		return nil, err
	}

	if apiKey.AccountId != account.Id {
		return nil, errors.New("ApiKey не принадлежит аккаунту")
	}

	return apiKey, nil
}

func (account Account) ApiKeysList() ([]ApiKey, error) {

	keyList, err := ApiKey{}.getList(account.Id)
	if err != nil {
		return nil, errors.New("Не удалось получить список")
	}

	return keyList, nil
}

func (account Account) ApiKeyUpdate(id uint, input map[string]interface{}) (*ApiKey, error) {
	apiKey, err := account.ApiKeyGet(id)
	if err != nil {
		return nil, err
	}

	if account.Id != apiKey.AccountId {
		return nil, utils.Error{Message: "Ключ принадлежит другому аккаунту"}
	}

	err = apiKey.update(input)

	return apiKey, err

}

func (account Account) ApiKeyDelete(id uint) error {

	apiKey, err := account.ApiKeyGet(id)
	if err != nil {
		return err
	}

	return apiKey.delete()
}



// ########### END OF ACCOUNT FUNCTIONAL ###########