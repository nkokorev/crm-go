package models

import (
	"errors"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

// Список системных функций обработки, которые можно вызвать в Observer
type HandlerItem struct {
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	Name		string 	`json:"name" gorm:"type:varchar(255);unique;not null;"`  // Вызов WebHook'а

	Code		string 	`json:"code" gorm:"type:varchar(255);unique;not null;"` // Имя функции 'WebHookCall'
	EntityType	string 	`json:"entity_type" gorm:"type:varchar(50);not null;"` // имя типа (таблицы): 'web_hooks'

	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:false;"` // Глобальный статус Observer



	Description string 	`json:"description" gorm:"type:text;"` // pgsql: text

	CreatedAt time.Time `json:"created_at"`
}

func (HandlerItem) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&HandlerItem{}); err != nil {log.Fatal(err)}
	// db.Model(&HandlerItem{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE handler_items ADD CONSTRAINT handler_items_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

	mainAccount, err := GetMainAccount()
	if err != nil {
		log.Fatal("Не удалось найти главный аккаунт для событий")
	}

	// HandlerItem
	eventHandlers := []HandlerItem {
		{Name:"Вызов WebHook'а", Code: "WebHookCall", EntityType: "web_hooks", Enabled: true, Description: "Вызов указанного WebHook'а"},
		{Name:"Запуск email-уведомления", Code: "EmailNotificationRun", EntityType: "email_notification", Enabled: true, Description: "Отправка электронного письма. Адресат выбирается в зависимости от настроек уведомления и события. Если объект пользователь - то на его email. При отсутствии email'а, запуск уведомления не произойдет."},
		{Name:"Запуск серии email", Code: "EmailQueueRun", EntityType: "email_queue", Enabled: true, Description: "Запуск автоматической серии писем. Адресат выбирается исходя из события. Если объект пользователь - то на его email. При отсутствии email'а, запуск серии не произойдет."},
	}
	for _,v := range eventHandlers {
		_, err = mainAccount.CreateEntity(&v)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func (handlerItem *HandlerItem) BeforeCreate(tx *gorm.DB) error {
	handlerItem.Id = 0
	return nil
}

// ############# Entity interface #############
func (handlerItem HandlerItem) GetId() uint           { return handlerItem.Id }
func (handlerItem *HandlerItem) setId(id uint)        { handlerItem.Id = id }
func (handlerItem *HandlerItem) setPublicId(id uint)  { }
func (handlerItem HandlerItem) GetAccountId() uint    { return handlerItem.AccountId }
func (handlerItem *HandlerItem) setAccountId(id uint) { handlerItem.AccountId = id }

// Всегда системный
func (handlerItem HandlerItem) SystemEntity() bool { return handlerItem.AccountId == 1 }

// ############# Entity interface #############

func (handlerItem HandlerItem) create() (Entity, error)  {

	_item := handlerItem
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	go EventListener{}.ReloadEventHandlers()

	return entity, nil
}
func (HandlerItem) get(id uint, preloads []string) (Entity, error) {

	var item HandlerItem

	err := item.GetPreloadDb(false, false, preloads).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
func (handlerItem *HandlerItem) load(preloads []string) error {
	if handlerItem.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить CartItem - не указан  Id"}
	}

	err := handlerItem.GetPreloadDb(false, false, preloads).First(handlerItem, handlerItem.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (*HandlerItem) loadByPublicId(preloads []string) error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}
func (HandlerItem) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return HandlerItem{}.getPaginationList(accountId,0,100,sortBy,"", nil,preload)
}
func (HandlerItem) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	obItems := make([]HandlerItem,0)
	var total int64

	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&HandlerItem{}).GetPreloadDb(false, false, preloads).
			Order(sortBy).Offset(offset).Limit(limit).
			Where("account_id IN (?)", []uint{1, accountId}).
			Find(&obItems, "name ILIKE ? OR description ILIKE ?",search,search).Error
		if err != nil {
			return nil, 0, err
		}

		// Определяем total
		err = (&HandlerItem{}).GetPreloadDb(false, false, nil).
			Where("account_id IN (?) AND name ILIKE ? OR description ILIKE ?", []uint{1, accountId}, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}


	} else {
		err := (&HandlerItem{}).GetPreloadDb(false, false, preloads).
			Order(sortBy).Offset(offset).Limit(limit).
			Where("account_id IN (?)", []uint{1, accountId}).
			Find(&obItems).Error
		if err != nil {
			return nil, 0, err
		}

		// Определяем total
		err = (&HandlerItem{}).GetPreloadDb(false, false, nil).Where("account_id IN (?)", []uint{1, accountId}).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(obItems))
	for i := range obItems {
		entities[i] = &obItems[i]
	}

	return entities, total, nil
}
func (handlerItem *HandlerItem) update(input map[string]interface{}, preloads []string) error {

	// delete(input,"amount")
	utils.FixInputHiddenVars(&input)
	/*if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id"}); err != nil {
		return err
	}*/

	if err := handlerItem.GetPreloadDb(false, false, nil).Where("id = ?", handlerItem.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	err := handlerItem.GetPreloadDb(false,false, preloads).First(handlerItem, handlerItem.Id).Error
	if err != nil {
		return err
	}

	go EventListener{}.ReloadEventHandlers()

	return nil
}
func (handlerItem *HandlerItem) delete () error {
	if err := handlerItem.GetPreloadDb(true,false,nil).Where("id = ?", handlerItem.Id).Delete(handlerItem).Error; err !=nil { return err}

	go EventListener{}.ReloadEventHandlers()

	return nil
}

func (handlerItem *HandlerItem) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(&handlerItem)
	} else {
		_db = _db.Model(&HandlerItem{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}