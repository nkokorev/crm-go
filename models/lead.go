package models

import (
	"time"
)

type Lead struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"accountId" gorm:"index,not null"` // аккаунт-владелец ключа

	Name string `json:"name" gorm:"type:varchar(255);default:'New api key';"` //
	Enabled bool `json:"enabled" gorm:"type:bool;default:true"` // активен ли ключ

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Lead) PgSqlCreate() {
	
	db.CreateTable(&Lead{})
	db.Model(&Lead{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	
}

/*func (lead Lead) getId() uint {
	return lead.ID
}
func (lead Lead) GetAccountId() uint {
	return lead.AccountID
}
func (lead *Lead) setAccountId(id uint) {
	lead.AccountID = id
}
func (Lead) getEntityName() string {
	return "Lead"
}

func (lead *Lead) BeforeCreate(scope *gorm.Scope) error {
	lead.ID = 0
	return nil
}



////////////

func (lead Lead) create() (*Entity, error)  {
	var newLead Entity = &lead
	
	if err := db.Create(newLead).Error; err != nil {
		return nil, err
	}

	return &newLead, nil
}

func (Lead) get(id uint) (*Entity, error) {

	// var entity Entity

	var lead Lead

	err := db.First(&lead, id).Error
	if err != nil {
		return nil, err
	}
	// entity = &lead

	return &lead, nil
}
*/
/*func (Lead) get(id uint) (interface{}, error) {

	// var entity Entity

	var lead Lead

	err := db.First(&lead, id).Error
	if err != nil {
		return nil, err
	}
	// entity = &lead

	return &lead, nil
}*/

/*func (lead Lead) create() (*Entity, error)  {
	var newLead = lead

	if err := db.Create(&newLead).Error; err != nil {
		return nil, err
	}

	var e Entity = &newLead
	return &e, nil
}*/

/*func (ApiKey) get(id uint) (*ApiKey, error) {

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

func (apiKey ApiKey) delete () error {
	return db.Model(ApiKey{}).Where("id = ?", apiKey.ID).Delete(apiKey).Error
}

func (apiKey *ApiKey) update(input interface{}) error {
	// return db.Model(apiKey).Omit("token", "account_id", "created_at", "updated_at").Select("Name", "Enabled").Updates(&input).Error
	return db.Model(apiKey).Select("Name", "Enabled").Updates(structs.Map(input)).Error

}*/

// ######## !!!! Все что выше покрыто тестами на прямую или косвено
