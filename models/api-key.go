package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"strings"
	"time"
)

type ApiKey struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	Token string `json:"token" gorm:"unique_index;not null;"` // ID
	AccountID uint `json:"accountId" gorm:"index,not null"` // аккаунт-владелец ключа

	Name string `json:"name" gorm:"type:varchar(255);default:'New api key';"` // имя ключа "Для сайта", "Для CRM"
	Enabled bool `json:"enabled" gorm:"type:bool;default:true"` // активен ли ключ

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (ApiKey) PgSqlCreate() {
	
	db.CreateTable(&ApiKey{})
	
	db.Exec("ALTER TABLE api_keys \n    ADD CONSTRAINT api_keys_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\n\n-- create unique index uix_api_keys_token_account_id ON api_keys (token,account_id);")
}

func (apiKey *ApiKey) BeforeCreate(scope *gorm.Scope) error {

	apiKey.ID = 0

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

func (apiKey ApiKey) delete () error {
	return db.Model(ApiKey{}).Where("id = ?", apiKey.ID).Delete(apiKey).Error
}

func (apiKey *ApiKey) update(input map[string]interface{}) error {
	// return db.Model(apiKey).Omit("token", "account_id", "created_at", "updated_at").Select("Name", "Enabled").Updates(&input).Error
	return db.Model(apiKey).Select("Name", "Enabled").Updates(input).Error

}

// ######## !!!! Все что выше покрыто тестами на прямую или косвено

// ########### ACCOUNT FUNCTIONAL ###########

func (account Account) ApiKeyCreate(input ApiKey) (*ApiKey, error) {
	input.AccountID = account.ID
	return ApiKey{}.create(input)
}

func (account Account) ApiKeyGet(id uint) (*ApiKey, error) {
	apiKey, err := ApiKey{}.get(id)
	if err != nil {
		return nil, err
	}

	if apiKey.AccountID != account.ID {
		return nil, errors.New("ApiKey не принадлежит аккаунту")
	}

	return apiKey, nil
}

func (account Account) ApiKeyGetByToken(token string) (*ApiKey, error) {
	apiKey, err := GetApiKeyByToken(token)
	if err != nil {
		return nil, err
	}

	if apiKey.AccountID != account.ID {
		return nil, errors.New("ApiKey не принадлежит аккаунту")
	}

	return apiKey, nil
}

func (account Account) ApiKeysList() ([]ApiKey, error) {

	keyList, err := ApiKey{}.getList(account.ID)
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

	if account.ID != apiKey.AccountID {
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