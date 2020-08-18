package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"log"
	"time"
)

// Список системных функций обработки, которые можно вызвать в Observer
type HandlerItem struct {
	Id     		uint   	`json:"id" gorm:"primary_key"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

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

func (obItem *HandlerItem) BeforeCreate(scope *gorm.Scope) error {
	obItem.Id = 0
	return nil
}

// ############# Entity interface #############
func (obItem HandlerItem) GetId() uint { return obItem.Id }
func (obItem *HandlerItem) setId(id uint) { obItem.Id = id }
func (obItem *HandlerItem) setPublicId(id uint) { }
func (obItem HandlerItem) GetAccountId() uint { return obItem.AccountId }
func (obItem *HandlerItem) setAccountId(id uint) { obItem.AccountId = id }

// Всегда системный
func (handlerItem HandlerItem) SystemEntity() bool { return handlerItem.AccountId == 1 }

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

	err := db.First(obItem,obItem.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (*HandlerItem) loadByPublicId() error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}
func (HandlerItem) getList(accountId uint, sortBy string) ([]Entity, uint, error) {

	obItems := make([]HandlerItem,0)
	var total uint

	err := db.Model(&HandlerItem{}).Limit(100).Order(sortBy).Where( "account_id IN (?)", []uint{1, accountId}).
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
	for i := range obItems {
		entities[i] = &obItems[i]
	}

	return entities, total, nil
}
func (HandlerItem) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{}) ([]Entity, uint, error) {

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
	for i := range obItems {
		entities[i] = &obItems[i]
	}

	return entities, total, nil
}
func (obItem *HandlerItem) update(input map[string]interface{}) error {
	if err := db.Set("gorm:association_autoupdate", false).Model(obItem).Omit("id", "account_id").
		Updates(input).Error; err != nil {return err}

	go EventListener{}.ReloadEventHandlers()

	return nil
}
func (obItem *HandlerItem) delete () error {
	return db.Model(HandlerItem{}).Where("id = ?", obItem.Id).Delete(obItem).Error
}