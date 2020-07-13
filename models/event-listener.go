package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
)

// Храним в БД список
type EventActions struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"-" gorm:"type:int;index;not null;"`

	// какой событие слушаем
	EventName string `json:"eventName"`

	// какое действие выполняем (имя функции)
	TargetId	uint 	`json:"targetId"`   // 1
	TargetName	string 	`json:"targetName"` //webhooks

	Priority int	`json:"priority"` // Приоритет выполнения
}

func (EventActions) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&EventActions{})
	db.Model(&WebHook{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")

}

func (el *EventActions) BeforeCreate(scope *gorm.Scope) error {
	el.ID = 0
	return nil
}

// выполняет функцию
func (el *EventActions) Fire() {
	switch el.TargetName {
	case "webHookRun":
		fmt.Println("Выполняем веб хук!")
	case "emailQueueRun":
		fmt.Println("Выполняем emailQueue!")
	default:
		fmt.Println("EventActions: Действие не определено!")
	}
}


// ############# Entity interface #############
func (el EventActions) getId() uint { return el.ID }
func (el *EventActions) setId(id uint) { el.ID = id }
func (el EventActions) GetAccountId() uint { return el.AccountID }
func (el *EventActions) setAccountId(id uint) { el.AccountID = id }
func (el EventActions) systemEntity() bool { return false }
// ############# END Of Entity interface #############


func (el EventActions) create() (Entity, error)  {
	var newItem Entity = &el

	if err := db.Create(newItem).Error; err != nil {
		return nil, err
	}

	return newItem, nil
}
func (EventActions) get(id uint) (Entity, error) {

	var el EventActions

	err := db.First(&el, id).Error
	if err != nil {
		return nil, err
	}
	return &el, nil
}
func (el *EventActions) load() error {

	err := db.First(el).Error
	if err != nil {
		return err
	}
	return nil
}
func (EventActions) getList(accountId uint) ([]EventActions, error) {

	els := make([]EventActions,0)

	err := db.Find(&els, "account_id = ?", accountId).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return els, nil
}

func (EventActions) getEnabledByName(accountId uint, eventName string) ([]EventActions, error) {

	els := make([]EventActions,0)

	err := db.Find(&els, "account_id = ? AND event_name = ?", accountId, eventName).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return els, nil
}
func (EventActions) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	els := make([]EventActions,0)
	var total uint

	err := db.Model(&EventActions{}).Limit(limit).Offset(offset).Order(sortBy).Find(&els, "account_id = ?", accountId).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Model(&EventActions{}).Where("account_id = ?", accountId).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(els))
	for i, v := range els {
		entities[i] = &v
	}

	return entities, total, nil
}
func (el *EventActions) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).Model(el).Omit("id", "account_id").Update(input).Error
}
func (el EventActions) delete () error {
	return db.Model(EventActions{}).Where("id = ?", el.ID).Delete(el).Error
}

