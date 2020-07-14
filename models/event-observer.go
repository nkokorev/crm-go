package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"reflect"
	"time"
)



type Observer struct {
	ID     		uint   	`json:"id" gorm:"primary_key"`
	AccountID 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	// какой событие слушаем
	EventName 	string 	`json:"eventName"`
	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"`

	// какое действие выполняем (имя функции)
	TargetId	uint 	`json:"targetId"`   // 1
	TargetName	string 	`json:"targetName"` //webhooks

	// Target		event.ListenerFunc 	`json:"target"` //webhooks

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
func (Observer) getAllAccountsList() ([]Observer, error) {

	observers := make([]Observer,0)

	err := db.Model(&Observer{}).Find(&observers).Error
	if err != nil {
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

func (Observer) getList(accountId uint, sortBy string) ([]Entity, uint, error) {

	observers := make([]Observer,0)
	var total uint

	err := db.Model(&Observer{}).Limit(1000).Order(sortBy).Where( "account_id = ?", accountId).
		Find(&observers).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, 0, err
	}

	// Определяем total
	err = db.Model(&Observer{}).Where("account_id = ?", accountId).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(observers))
	for i,_ := range observers {
		entities[i] = &observers[i]
	}

	return entities, total, nil
}
func (Observer) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	observers := make([]Observer,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&Observer{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&observers, "event_name ILIKE ? OR target_name ILIKE ?", search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Observer{}).
			Where("account_id = ? AND event_name ILIKE ? OR target_name ILIKE ?", accountId, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&Observer{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&observers).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Observer{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(observers))
	for i,_ := range observers {
		entities[i] = &observers[i]
	}

	return entities, total, nil
}
func (observer *Observer) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).Model(observer).Omit("id", "account_id").Update(input).Error
}
func (observer Observer) delete () error {
	return db.Model(Observer{}).Where("id = ?", observer.ID).Delete(observer).Error
}

// Нужна функция ReloadEventHandler(e)
func (Observer) Registration() error {

	eventListeners, err := Observer{}.getAllAccountsList()
	if err != nil {
		return utils.Error{Message: "Не удалось загрузить EventHandlers!"}
	}

	for _,v := range eventListeners {
		event.On(v.EventName, v, v.Priority)
	}

	return nil
}

func (Observer) ReloadEventHandler() error {
	em := event.DefaultEM
	em.Clear()

	return Observer{}.Registration()
}

// для интерфейса event.Listener - функция обработчик для каждого события
// Она вызывается в цепочке первой, а затем уже соответствующая функция target из Observer (см. ниже)
func (observer Observer) Handle(e event.Event) error {

	// 1. Получаем метод обработки по имени Target
	m := reflect.ValueOf(observer).MethodByName(observer.TargetName)
	if m.IsNil() {
		e.Abort(true)
		return utils.Error{Message: fmt.Sprintf("Observer Handle is nill: %v", observer.TargetName)}
	}

	// 2. Преобразуем метод, чтобы его можно было вызвать от объекта Event
	target, ok := m.Interface().(func(e event.Event) error)
	if !ok {
		e.Abort(true)
		return utils.Error{Message: fmt.Sprintf("Observer mCallable !ok: %v", observer.TargetName)}
	}

	// 3. Вызываем Target-метод с объектом Event
	if err := target(e); err != nil {
		e.Abort(true)
		return err
	}

	return nil
}

// ########################################################
// Функции, которые могут быть вызваны для обработки событий типа Event
// Для добавления функции обработчика нужно:
// 1. Создать функцию ниже func (Observer) FName(event.Event) error
// 2. Создать запись в таблице ObserverList, добавим описание и назначение функции

// #############   Event Handlers   #############
func (observer Observer) EmailQueueRun(e event.Event) error {
	fmt.Printf("Запуск серии писем, данные: %v\n", e.Data())
	// fmt.Println("Observer: ", observer) // контекст серии писем, какой именно и т.д.
	// e.Set("result", "OK") // возможность записать в событие какие-то данные для других обработчиков..
	return nil
}
func (observer Observer) WebHookCall(e event.Event) error {
	fmt.Printf("Вызов вебхука, данные: %v\n", e.Data())
	// fmt.Println("Observer: ", observer) // контекст вебхука, какой именно и т.д.
	// e.Set("result", "OK")
	return nil
}
// #############   END Of Event Handlers   #############
