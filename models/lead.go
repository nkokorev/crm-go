package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
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

// ############# Entity interface #############
func (lead Lead) getId() uint           { return lead.ID }
func (lead *Lead) setId(id uint)        { lead.ID = id }
func (lead Lead) GetAccountId() uint    { return lead.AccountID }
func (lead *Lead) setAccountId(id uint) { lead.AccountID = id }
func (Lead) systemEntity() bool { return false }
// ############# Entity interface #############


// ###### GORM Functional #######
func (Lead) TableName() string { return "leads" }
func (lead *Lead) BeforeCreate(scope *gorm.Scope) error {
	lead.ID = 0
	return nil
}
// ###### End of GORM Functional #######

// ############# CRUD Entity interface #############

func (lead Lead) create() (Entity, error)  {
	var newItem Entity = &lead

	if err := db.Create(newItem).Error; err != nil {
		return nil, err
	}

	return newItem, nil
}

func (Lead) get(id uint) (Entity, error) {

	var lead Lead

	err := db.First(&lead, id).Error
	if err != nil {
		return nil, err
	}
	return &lead, nil
}

func (lead *Lead) load() error {

	err := db.First(lead).Error
	if err != nil {
		return err
	}
	return nil
}

func (Lead) getPaginationList(accountId uint, offset, limit int, order string, search string) ([]Entity, uint, error) {

	delivers := make([]Lead,0)
	var total uint

	err := db.Model(&Lead{}).Limit(limit).Offset(offset).Order(order).Find(&delivers, "account_id = ?", accountId).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Model(&Lead{}).Where("account_id = ?", accountId).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(delivers))
	for i, v := range delivers {
		entities[i] = &v
	}

	return entities, total, nil
}

func (lead *Lead) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).Model(lead).Omit("id", "account_id").Update(input).Error
}

func (lead Lead) delete () error {
	return db.Model(Lead{}).Where("id = ?", lead.ID).Delete(lead).Error
}

// ########## End of CRUD Entity interface ###########