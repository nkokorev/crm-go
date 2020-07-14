package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"reflect"
)

// Храним в БД список
type EventHandler struct {
	ID     		uint   	`json:"id" gorm:"primary_key"`
	AccountID 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	// какой событие слушаем
	EventName 	string 	`json:"eventName"`
	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"`

	// какое действие выполняем (имя функции)
	TargetId	uint 	`json:"targetId"`   // 1
	TargetName	string 	`json:"targetName"` //webhooks

	Priority 	int		`json:"priority" gorm:"type:int;default:0"` // Приоритет выполнения, по умолчанию 0 - Normal
}



func (EventHandler) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&EventHandler{})
	db.Model(&WebHook{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")

}

func (el *EventHandler) BeforeCreate(scope *gorm.Scope) error {
	el.ID = 0
	return nil
}

// ############# Entity interface #############
func (el EventHandler) getId() uint { return el.ID }
func (el *EventHandler) setId(id uint) { el.ID = id }
func (el EventHandler) GetAccountId() uint { return el.AccountID }
func (el *EventHandler) setAccountId(id uint) { el.AccountID = id }
func (el EventHandler) systemEntity() bool { return false }
// ############# END Of Entity interface #############


func (el EventHandler) create() (Entity, error)  {
	var newItem Entity = &el

	if err := db.Create(newItem).Error; err != nil {
		return nil, err
	}

	return newItem, nil
}
func (EventHandler) get(id uint) (Entity, error) {

	var el EventHandler

	err := db.First(&el, id).Error
	if err != nil {
		return nil, err
	}
	return &el, nil
}
func (el *EventHandler) load() error {

	err := db.First(el).Error
	if err != nil {
		return err
	}
	return nil
}
func (EventHandler) getFullList() ([]EventHandler, error) {

	els := make([]EventHandler,0)

	err := db.Model(&EventHandler{}).Find(&els).Error
	if err != nil {
		return nil, err
	}
	
	return els, nil
}
func (EventHandler) getList(accountId uint) ([]EventHandler, error) {

	els := make([]EventHandler,0)

	err := db.Find(&els, "account_id = ?", accountId).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return els, nil
}

func (EventHandler) getEnabledByName(accountId uint, eventName string) ([]EventHandler, error) {

	els := make([]EventHandler,0)

	err := db.Find(&els, "account_id = ? AND event_name = ?", accountId, eventName).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return els, nil
}
func (EventHandler) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	els := make([]EventHandler,0)
	var total uint

	err := db.Model(&EventHandler{}).Limit(limit).Offset(offset).Order(sortBy).Find(&els, "account_id = ?", accountId).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Model(&EventHandler{}).Where("account_id = ?", accountId).Count(&total).Error
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
func (el *EventHandler) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).Model(el).Omit("id", "account_id").Update(input).Error
}
func (el EventHandler) delete () error {
	return db.Model(EventHandler{}).Where("id = ?", el.ID).Delete(el).Error
}

// ########################################################

// функция обработчик для каждого события
func (eh EventHandler) Handle(e event.Event) error {

	// 1. Получаем метод обработки по имени Target
	m := reflect.ValueOf(eh).MethodByName(eh.TargetName)
	if m.IsNil() {
		e.Abort(true)
		return utils.Error{Message: fmt.Sprintf("EventHandler Handle is nill: %v", eh.TargetName)}
	}

	// 2. Преобразуем метод, чтобы его можно было вызвать от объекта Event
	target, ok := m.Interface().(func(e event.Event) error)
	if !ok {
		e.Abort(true)
		return utils.Error{Message: fmt.Sprintf("EventHandler mCallable !ok: %v", eh.TargetName)}
	}

	// 3. Вызываем Target-метод с объектом Event
	if err := target(e); err != nil {
		e.Abort(true)
		return err
	}

	return nil
}

// #############   Event Handlers   #############
func (eh EventHandler) EmailQueueRun(e event.Event) error {
	fmt.Printf("Запуск серии писем, данные: %v\n", e.Data())
	fmt.Println("EventHandler: ", eh) // контекст серии писем, какой именно и т.д.
	// e.Set("result", "OK") // возможность записать в событие какие-то данные для других обработчиков..
	return nil
}
func (eh EventHandler) WebHookCall(e event.Event) error {
	fmt.Printf("Вызов вебхука, данные: %v\n", e.Data())
	fmt.Println("EventHandler: ", eh) // контекст вебхука, какой именно и т.д.
	// e.Set("result", "OK")
	return nil
}
// #############   END Of Event Handlers   #############
