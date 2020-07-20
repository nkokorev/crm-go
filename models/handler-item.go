package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

// Список системных функций обработки, которые можно вызвать в Observer
type HandlerItem struct {
	ID     		uint   	`json:"id" gorm:"primary_key"`
	AccountID 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	Name		string 	`json:"name" gorm:"type:varchar(255);unique;not null;"`  // Вызов WebHook'а

	Code		string 	`json:"code" gorm:"type:varchar(255);unique;not null;"` // Имя функции 'WebHookCall'
	EntityType	string 	`json:"entityType" gorm:"type:varchar(50);not null;"` // имя типа (таблицы): 'web_hooks'

	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:false;"` // Глобальный статус Observer



	Description string 	`json:"description" gorm:"type:text;"` // pgsql: text

	CreatedAt time.Time `json:"createdAt"`
}


func (HandlerItem) PgSqlCreate() {
	db.CreateTable(&HandlerItem{})
	db.Model(&HandlerItem{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}

func (obItem *HandlerItem) BeforeCreate(scope *gorm.Scope) error {
	obItem.ID = 0
	return nil
}

// ############# Entity interface #############
func (obItem HandlerItem) getId() uint { return obItem.ID }
func (obItem *HandlerItem) setId(id uint) { obItem.ID = id }
func (obItem HandlerItem) GetAccountId() uint { return obItem.AccountID }
func (obItem *HandlerItem) setAccountId(id uint) { obItem.AccountID = id }
func (HandlerItem) systemEntity() bool { return true }

// ############# Entity interface #############


func (handlerItem HandlerItem) create() (Entity, error)  {

	hi := handlerItem

	if err := db.Create(&hi).Error; err != nil {
		return nil, err
	}
	var entity Entity = &hi

	return entity, nil
}

func (HandlerItem) get(id uint) (Entity, error) {

	var obItem HandlerItem

	err := db.First(&obItem, id).Error
	if err != nil {
		return nil, err
	}
	return &obItem, nil
}

func (obItem *HandlerItem) load() error {

	err := db.First(obItem).Error
	if err != nil {
		return err
	}
	return nil
}

func (HandlerItem) getList(accountId uint, sortBy string) ([]Entity, uint, error) {

	obItems := make([]HandlerItem,0)
	var total uint

	err := db.Model(&HandlerItem{}).Limit(1000).Order(sortBy).Where( "account_id IN (?)", []uint{1, accountId}).
		Find(&obItems).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, 0, err
	}

	// Определяем total
	err = db.Model(&HandlerItem{}).Where( "account_id IN (?)", []uint{1, accountId}).Count(&total).Error
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

func (HandlerItem) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	obItems := make([]HandlerItem,0)
	var total uint

	if len(search) > 0 {

		search = "%"+search+"%"

		err := db.Model(&HandlerItem{}).
			Order(sortBy).Offset(offset).Limit(limit).
			Where("account_id IN (?)", []uint{1, accountId}).
			Find(&obItems, "name ILIKE ? OR description ILIKE ?",search,search).Error
		if err != nil {
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&HandlerItem{}).
			Where("account_id IN (?) AND name ILIKE ? OR description ILIKE ?", []uint{1, accountId}, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}


	} else {
		err := db.Model(&HandlerItem{}).
			Order(sortBy).Offset(offset).Limit(limit).
			Where("account_id IN (?)", []uint{1, accountId}).
			Find(&obItems).Error
		if err != nil {
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&HandlerItem{}).Where("account_id IN (?)", []uint{1, accountId}).Count(&total).Error
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

func (obItem *HandlerItem) update(input map[string]interface{}) error {
	if err := db.Set("gorm:association_autoupdate", false).Model(obItem).Omit("id", "account_id").Update(input).Error; err != nil {return err}

	go EventListener{}.ReloadEventHandlers()

	return nil
}

func (obItem HandlerItem) delete () error {
	return db.Model(HandlerItem{}).Where("id = ?", obItem.ID).Delete(obItem).Error
}