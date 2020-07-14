package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type ObserverItem struct {
	ID     		uint   	`json:"id" gorm:"primary_key"`
	AccountID 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	Name		string 	`json:"name" gorm:"type:varchar(255);"`
	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // Глобальный статус Observer

	Description string 	`json:"description" gorm:"type:text;"` // pgsql: text

	CreatedAt time.Time `json:"createdAt"`
}


func (ObserverItem) PgSqlCreate() {
	db.CreateTable(&ObserverItem{})
	db.Model(&ObserverItem{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}

func (obItem *ObserverItem) BeforeCreate(scope *gorm.Scope) error {
	obItem.ID = 0
	return nil
}

// ############# Entity interface #############
func (obItem ObserverItem) getId() uint { return obItem.ID }
func (obItem *ObserverItem) setId(id uint) { obItem.ID = id }
func (obItem ObserverItem) GetAccountId() uint { return obItem.AccountID }
func (obItem *ObserverItem) setAccountId(id uint) { obItem.AccountID = id }
func (ObserverItem) systemEntity() bool { return true }

// ############# Entity interface #############


func (obItem ObserverItem) create() (Entity, error)  {
	var newItem Entity = &obItem

	if err := db.Create(newItem).Error; err != nil {
		return nil, err
	}

	return newItem, nil
}

func (ObserverItem) get(id uint) (Entity, error) {

	var obItem ObserverItem

	err := db.First(&obItem, id).Error
	if err != nil {
		return nil, err
	}
	return &obItem, nil
}

func (obItem *ObserverItem) load() error {

	err := db.First(obItem).Error
	if err != nil {
		return err
	}
	return nil
}

func (ObserverItem) getList(accountId uint, sortBy string) ([]Entity, uint, error) {

	obItems := make([]ObserverItem,0)
	var total uint

	err := db.Model(&ObserverItem{}).Limit(1000).Order(sortBy).Where( "account_id IN (?)", []uint{1, accountId}).
		Find(&obItems).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, 0, err
	}

	// Определяем total
	err = db.Model(&ObserverItem{}).Where( "account_id IN (?)", []uint{1, accountId}).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(obItems))
	for i,_ := range obItems {
		entities[i] = &obItems[i]
	}

	return entities, total, nil
}

func (ObserverItem) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	obItems := make([]ObserverItem,0)
	var total uint

	if len(search) > 0 {

		search = "%"+search+"%"

		err := db.Model(&ObserverItem{}).
			Order(sortBy).Offset(offset).Limit(limit).
			Where("account_id IN (?)", []uint{1, accountId}).
			Find(&obItems, "name ILIKE ? OR description ILIKE ?",search,search).Error
		if err != nil {
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&ObserverItem{}).
			Where("account_id IN (?) AND name ILIKE ? OR description ILIKE ?", []uint{1, accountId}, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}


	} else {
		err := db.Model(&ObserverItem{}).
			Order(sortBy).Offset(offset).Limit(limit).
			Where("account_id IN (?)", []uint{1, accountId}).
			Find(&obItems).Error
		if err != nil {
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&ObserverItem{}).Where("account_id IN (?)", []uint{1, accountId}).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(obItems))
	for i,_ := range obItems {
		entities[i] = &obItems[i]
	}

	return entities, total, nil
}

func (obItem *ObserverItem) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).Model(obItem).Omit("id", "account_id").Update(input).Error
}

func (obItem ObserverItem) delete () error {
	return db.Model(ObserverItem{}).Where("id = ?", obItem.ID).Delete(obItem).Error
}