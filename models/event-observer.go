package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

// Храним в БД список
type Observer struct {
	ID     		uint   	`json:"id" gorm:"primary_key"`
	AccountID 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	// какой событие слушаем
	EventName 	string 	`json:"eventName"`
	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"`

	// какое действие выполняем (имя функции)
	TargetId	uint 	`json:"targetId"`   // 1
	TargetName	string 	`json:"targetName"` //webhooks

	Priority 	int		`json:"priority" gorm:"type:int;default:0"` // Приоритет выполнения, по умолчанию 0 - Normal

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Observer) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&Observer{})
	db.Model(&WebHook{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")

}

func (observer *Observer) BeforeCreate(scope *gorm.Scope) error {
	observer.ID = 0
	return nil
}

// ############# Entity interface #############
func (observer Observer) getId() uint { return observer.ID }
func (observer *Observer) setId(id uint) { observer.ID = id }
func (observer Observer) GetAccountId() uint { return observer.AccountID }
func (observer *Observer) setAccountId(id uint) { observer.AccountID = id }
func (observer Observer) systemEntity() bool { return false }
// ############# END Of Entity interface #############

func (observer Observer) create() (Entity, error)  {
	var newItem Entity = &observer

	if err := db.Create(newItem).Error; err != nil {
		return nil, err
	}

	return newItem, nil
}
func (Observer) get(id uint) (Entity, error) {

	var observer Observer

	err := db.First(&observer, id).Error
	if err != nil {
		return nil, err
	}
	return &observer, nil
}
func (observer *Observer) load() error {

	err := db.First(observer).Error
	if err != nil {
		return err
	}
	return nil
}
func (Observer) getFullList() ([]Observer, error) {

	observers := make([]Observer,0)

	err := db.Model(&Observer{}).Find(&observers).Error
	if err != nil {
		return nil, err
	}
	
	return observers, nil
}
func (Observer) getList(accountId uint) ([]Observer, error) {

	observers := make([]Observer,0)

	err := db.Find(&observers, "account_id = ?", accountId).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return observers, nil
}

func (Observer) getEnabledByName(accountId uint, eventName string) ([]Observer, error) {

	observers := make([]Observer,0)

	err := db.Find(&observers, "account_id = ? AND event_name = ?", accountId, eventName).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return observers, nil
}
func (Observer) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	eventHandlers := make([]Observer,0)
	var total uint

	err := db.Model(&Observer{}).Limit(limit).Offset(offset).Order(sortBy).Find(&eventHandlers, "account_id = ?", accountId).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Model(&Observer{}).Where("account_id = ?", accountId).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(eventHandlers))
	for i, _ := range eventHandlers {
		entities[i] = &eventHandlers[i]
	}

	return entities, total, nil
}
func (observer *Observer) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).Model(observer).Omit("id", "account_id").Update(input).Error
}
func (observer Observer) delete () error {
	return db.Model(Observer{}).Where("id = ?", observer.ID).Delete(observer).Error
}
