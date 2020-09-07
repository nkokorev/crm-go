package models

import (
	"errors"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

// Связывает Event с собственными функциями Handler
type EventListener struct {
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	Name        string   `json:"name" gorm:"type:varchar(255);"` // для чего ?
	
	EventId		uint 	`json:"event_id" gorm:"type:int;"` // аналог EventName
	HandlerId	uint 	`json:"handler_id" gorm:"type:int;"` // аналог EventName

	EntityId	uint 	`json:"entity_id" gorm:"type:int;"` // Id целевого entity
	// EntityType	string 	`json:"entity_type" gorm:"type:varchar(50);default:''"` // таблица / Объект

	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:false"`

	Priority 	int		`json:"priority" gorm:"type:int;default:0"` // Приоритет выполнения, по умолчанию 0 - Normal

	Event 		EventItem 		`json:"event"`
	Handler 	HandlerItem		`json:"handler"`       // gorm:"preload:true"

	// TargetName string `json:"-" gorm:"-"`// Имя функции, которую вызывают локально

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (EventListener) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	if err := db.Migrator().CreateTable(&EventListener{});err != nil {log.Fatal(err)}
	// db.Model(&EventListener{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&EventListener{}).AddForeignKey("event_id", "event_items(id)", "CASCADE", "CASCADE")
	// db.Model(&EventListener{}).AddForeignKey("handler_id", "handler_items(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE event_listeners " +
		"ADD CONSTRAINT event_listeners_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT event_listeners_event_id_fkey FOREIGN KEY (event_id) REFERENCES event_items(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT event_listeners_handler_id_fkey FOREIGN KEY (handler_id) REFERENCES handler_items(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (eventListener *EventListener) BeforeCreate(tx *gorm.DB) error {
	eventListener.Id = 0
	return nil
}

// ############# Entity interface #############
func (eventListener EventListener) GetId() uint           { return eventListener.Id }
func (eventListener *EventListener) setId(id uint)        { eventListener.Id = id }
func (eventListener *EventListener) setPublicId(id uint)  { eventListener.Id = id }
func (eventListener EventListener) GetAccountId() uint    { return eventListener.AccountId }
func (eventListener *EventListener) setAccountId(id uint) { eventListener.AccountId = id }
func (eventListener EventListener) SystemEntity() bool    { return false }
// ############# END Of Entity interface #############

func (eventListener EventListener) create() (Entity, error)  {
	_item := eventListener
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}
func (EventListener) get(id uint, preloads []string) (Entity, error) {

	var item EventListener

	err := item.GetPreloadDb(false, false, preloads).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
func (eventListener *EventListener) load(preloads []string) error {

	if eventListener.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить EventListener - не указан  Id"}
	}

	err := eventListener.GetPreloadDb(false, false, preloads).First(eventListener, eventListener.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (*EventListener) loadByPublicId(preloads []string) error {
	return errors.New("Нет возможности загрузить объект по Public Id")
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
func (EventListener) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return EventListener{}.getPaginationList(accountId,0,200,sortBy,"", nil, preload)
}
func (EventListener) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	type EventListenerSearch struct {
		EventListener
		EventItem
	}
	// eventListeners := make([]EventListenerSearch,0)
	eventListeners := make([]EventListener,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		// err := db.Table("event_listeners").Model(&EventListener{}).
		err := (&EventListener{}).GetPreloadDb(false, false, preloads).Table("event_listeners").
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

		err := (&EventListener{}).GetPreloadDb(false, false, preloads).
			Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&eventListeners).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&EventListener{}).GetPreloadDb(false, false, nil).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(eventListeners))
	for i := range eventListeners {
		// entities[i] = &eventListeners[i].EventListener
		entities[i] = &eventListeners[i]
	}
	
	return entities, total, nil
}
func (eventListener *EventListener) update(input map[string]interface{}, preloads []string) error {

	delete(input,"event")
	delete(input,"handler")

	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"entity_id","event_id","handler_id","priority"}); err != nil {
		return err
	}

	// фиксируем состояние ДО обновления
	if err := db.Model(&EventListener{}).Where(" id = ?", eventListener.Id).Omit("id","account_id","created_at", "updated_at", "event", "handler").
		Updates(input).Error; err != nil {
			return err
	}

	go EventListener{}.ReloadEventHandlers()

	return nil
}
func (eventListener *EventListener) delete () error {
	if err := db.Model(EventListener{}).Where("id = ?", eventListener.Id).Delete(eventListener).Error; err != nil {return err}

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
		return
	}
	em := event.DefaultEM
	em.AddListener(eventListener.Event.Code, &eventListener, eventListener.Priority)
}

func (eventListener *EventListener) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(&eventListener)
	} else {
		_db = _db.Model(&EventListener{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"Event","Handler"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}

// для интерфейса event.Listener - функция обработчик для каждого события
// Она вызывается в цепочке первой, а затем уже соответствующая функция target из EventListener (см. ниже)


// ########################################################
// Функции, которые могут быть вызваны для обработки событий типа Event
// Для добавления функции обработчика нужно:
// 1. Создать функцию ниже func (EventListener) FName(event.Event) error
// 2. Создать запись в таблице ObserverList, добавим описание и назначение функции


