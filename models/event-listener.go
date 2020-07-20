package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"log"
	"time"
)

// Связывает Event с собственными функциями Handler
type EventListener struct {
	ID     		uint   	`json:"id" gorm:"primary_key"`
	AccountID 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	Name        string   `json:"name" gorm:"type:varchar(255);"` // для чего ?
	
	EventID		uint `json:"eventId" gorm:"type:int;"` // аналог EventName
	HandlerID	uint `json:"handlerId" gorm:"type:int;"` // аналог EventName

	EntityId	uint 	`json:"entityId" gorm:"type:int;"` // ID целевого entity
	// EntityType	string 	`json:"entityType" gorm:"type:varchar(50);default:''"` // таблица / Объект

	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:false"`

	Priority 	int		`json:"priority" gorm:"type:int;default:0"` // Приоритет выполнения, по умолчанию 0 - Normal

	Event 		EventItem 	`json:"event"  gorm:"preload:true"`
	Handler 	HandlerItem		`json:"handler"  gorm:"preload:true"`       // gorm:"preload:true"

	// TargetName string `json:"-" gorm:"-"`// Имя функции, которую вызывают локально

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (EventListener) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&EventListener{})
	db.Model(&EventListener{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&EventListener{}).AddForeignKey("event_id", "event_items(id)", "CASCADE", "CASCADE")
	db.Model(&EventListener{}).AddForeignKey("handler_id", "handler_items(id)", "CASCADE", "CASCADE")

}

func (eventListener *EventListener) BeforeCreate(scope *gorm.Scope) error {
	eventListener.ID = 0
	return nil
}

// ############# Entity interface #############
func (eventListener EventListener) getId() uint { return eventListener.ID }
func (eventListener *EventListener) setId(id uint) { eventListener.ID = id }
func (eventListener EventListener) GetAccountId() uint { return eventListener.AccountID }
func (eventListener *EventListener) setAccountId(id uint) { eventListener.AccountID = id }
func (eventListener EventListener) systemEntity() bool { return false }
// ############# END Of Entity interface #############

func (eventListener EventListener) create() (Entity, error)  {
	var newItem Entity = &eventListener

	if err := db.Create(newItem).Preload("Event").Preload("Handler").First(newItem).Error; err != nil {
		return nil, err
	}

	go eventListener.LoadListener()

	return newItem, nil
}
func (EventListener) get(id uint) (Entity, error) {

	var eventListener EventListener

	err := db.Preload("Event").Preload("Handler").First(&eventListener, id).Error
	if err != nil {
		return nil, err
	}
	return &eventListener, nil
}
func (eventListener *EventListener) load() error {

	err := db.Preload("Event").Preload("Handler").First(eventListener).Error
	if err != nil {
		return err
	}
	return nil
}

// функция для загрузки слушателей в шину Events <-
func (EventListener) getAllAccountsList() ([]EventListener, error) {

	eventListeners := make([]EventListener,0)

	err := db.Model(&EventListener{}).Preload("Event").Preload("Handler").Find(&eventListeners).Error
	if err != nil {
		return nil, err
	}
	
	return eventListeners, nil
}
func (EventListener) getEnabledByName(accountId uint, eventName string) ([]EventListener, error) {

	eventListeners := make([]EventListener,0)

	err := db.Find(&eventListeners, "account_id = ? AND event_name = ?", accountId, eventName).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return eventListeners, nil
}
func (EventListener) getList(accountId uint, sortBy string) ([]Entity, uint, error) {

	eventListeners := make([]EventListener,0)
	var total uint

	err := db.Model(&EventListener{}).Preload("Event").Preload("Handler").Limit(1000).Order(sortBy).Where( "account_id = ?", accountId).
		Find(&eventListeners).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, 0, err
	}

	// Определяем total
	err = db.Model(&EventListener{}).Where("account_id = ?", accountId).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(eventListeners))
	for i,_ := range eventListeners {
		entities[i] = &eventListeners[i]
	}

	return entities, total, nil
}
func (EventListener) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	type EventListenerSearch struct {
		EventListener
		EventItem
	}
	eventListeners := make([]EventListenerSearch,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Table("event_listeners").Model(&EventListener{}).
			Joins("LEFT JOIN event_items ON event_items.id = event_listeners.event_id").Select("event_items.name as e_name, event_listeners.*").
			Limit(limit).Offset(offset).Order(sortBy).Where( "event_listeners.account_id = ?", accountId).Preload("Event").Preload("Handler").
			Find(&eventListeners, "event_listeners.name ILIKE ? OR event_items.name ILIKE ?", search,search).Error
		
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Table("event_listeners").
			Joins("LEFT JOIN event_items ON event_items.id = event_listeners.event_id").Select("event_items.name as e_name, event_items.description as e_desc, event_listeners.*").
			Where( "event_listeners.account_id = ? AND event_listeners.name ILIKE ? OR event_items.name ILIKE ?", accountId, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Table("event_listeners").Model(&EventListener{}).Preload("Event").Preload("Handler").Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&eventListeners).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Table("event_listeners").Model(&EventListener{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(eventListeners))
	for i := range eventListeners {
		// entities[i] = &eventListeners[i]
		entities[i] = &eventListeners[i].EventListener
	}
	
	return entities, total, nil
}
func (eventListener *EventListener) update(input map[string]interface{}) error {

	// фиксируем состояние ДО обновления
	if err := db.Set("gorm:association_autoupdate", false).
		Model(eventListener).Omit("id","account_id","created_at", "updated_at", "event", "handler").Update(input).Error; err != nil {
			return err
	}

	go EventListener{}.ReloadEventHandlers()

	return nil
}
func (eventListener EventListener) delete () error {
	if err := db.Model(EventListener{}).Where("id = ?", eventListener.ID).Delete(eventListener).Error; err != nil {return err}

	go EventListener{}.ReloadEventHandlers()

	return nil
}

// Нужна функция ReloadEventHandler(e)
func (EventListener) Registration() error {

	eventListeners, err := EventListener{}.getAllAccountsList()
	if err != nil {
		return utils.Error{Message: "Не удалось загрузить EventHandlers!"}
	}

	// fmt.Println("eventListeners")

	for i,v := range eventListeners {
		if v.Enabled && eventListeners[i].Event.Enabled && eventListeners[i].Handler.Enabled {
			// fmt.Println("Event listener: ", v.Event.Name, " - ", v.Handler.Name)
			eventListeners[i].LoadListener()
			// event.On(v.Event.Name, Handler{TargetName: v.Handler.Name}, v.Priority)
		}
	}

	return nil
}

func (EventListener) ReloadEventHandlers() error {
	em := event.DefaultEM
	em.Clear()

	return EventListener{}.Registration()
}

func (eventListener EventListener) LoadListener() {
	// fmt.Println(": ", eventListener.Event)
	if eventListener.Event.Code == "" || eventListener.Handler.Code == "" {
		log.Println("LoadListener is empty name")
		// log.Println("LoadListener is empty name", eventListener.Event, eventListener.Handler)
		return
	}
	em := event.DefaultEM
	em.AddListener(eventListener.Event.Code, &eventListener, eventListener.Priority)
}

// для интерфейса event.Listener - функция обработчик для каждого события
// Она вызывается в цепочке первой, а затем уже соответствующая функция target из EventListener (см. ниже)


// ########################################################
// Функции, которые могут быть вызваны для обработки событий типа Event
// Для добавления функции обработчика нужно:
// 1. Создать функцию ниже func (EventListener) FName(event.Event) error
// 2. Создать запись в таблице ObserverList, добавим описание и назначение функции


