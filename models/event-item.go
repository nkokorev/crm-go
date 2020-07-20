package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

// Список событий, на которые можно навесить обработчик EventHandler
type EventItem struct {
	ID     		uint   	`json:"id" gorm:"primary_key"`
	AccountID 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	Name		string 	`json:"name" gorm:"type:varchar(255);unique;not null;"`  // 'Пользователь создан'
	Code		string 	`json:"code" gorm:"type:varchar(255);unique;not null;"`  // 'UserCreated'
	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:false;"` // Глобальный статус события (вызывать ли его или нет)

	Description string 	`json:"description" gorm:"type:text;"` // pgsql: text

	CreatedAt time.Time `json:"createdAt"`
}

func (EventItem) PgSqlCreate() {
	db.CreateTable(&EventItem{})
	db.Model(&EventItem{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}

func (eventItem *EventItem) BeforeCreate(scope *gorm.Scope) error {
	eventItem.ID = 0
	return nil
}

// ############# Entity interface #############
func (eventItem EventItem) getId() uint { return eventItem.ID }
func (eventItem *EventItem) setId(id uint) { eventItem.ID = id }
func (eventItem EventItem) GetAccountId() uint { return eventItem.AccountID }
func (eventItem *EventItem) setAccountId(id uint) { eventItem.AccountID = id }
func (EventItem) systemEntity() bool { return true }

// ############# Entity interface #############


func (eventItem EventItem) create() (Entity, error)  {

	ei := eventItem

	if err := db.Create(&ei).Error; err != nil {
		return nil, err
	}
	var entity Entity = &ei

	return entity, nil
}

func (EventItem) get(id uint) (Entity, error) {

	var eventItem EventItem

	err := db.First(&eventItem, id).Error
	if err != nil {
		return nil, err
	}
	return &eventItem, nil
}

func (eventItem *EventItem) load() error {

	err := db.First(eventItem).Error
	if err != nil {
		return err
	}
	return nil
}

func (EventItem) getList(accountId uint, sortBy string) ([]Entity, uint, error) {

	eventItems := make([]EventItem,0)
	var total uint

	err := db.Model(&EventItem{}).Limit(1000).Order(sortBy).Where( "account_id IN (?)", []uint{1, accountId}).
		Find(&eventItems).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, 0, err
	}

	// Определяем total
	err = db.Model(&EventItem{}).Where( "account_id IN (?)", []uint{1, accountId}).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(eventItems))
	for i,_ := range eventItems {
		entities[i] = &eventItems[i]
	}

	return entities, total, nil
}

func (EventItem) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	eventItems := make([]EventItem,0)
	var total uint

	if len(search) > 0 {

		search = "%"+search+"%"

		err := db.Model(&EventItem{}).
			Order(sortBy).Offset(offset).Limit(limit).
			Where("account_id IN (?)", []uint{1, accountId}).
			Find(&eventItems, "name ILIKE ? OR description ILIKE ?",search,search).Error
		if err != nil {
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&EventItem{}).
			Where("account_id IN (?) AND name ILIKE ? OR description ILIKE ?", []uint{1, accountId}, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}


	} else {
		err := db.Model(&EventItem{}).
			Order(sortBy).Offset(offset).Limit(limit).
			Where("account_id IN (?)", []uint{1, accountId}).
			Find(&eventItems).Error
		if err != nil {
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&EventItem{}).Where("account_id IN (?)", []uint{1, accountId}).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(eventItems))
	for i,_ := range eventItems {
		entities[i] = &eventItems[i]
	}

	return entities, total, nil
}

func (eventItem *EventItem) update(input map[string]interface{}) error {
	if err := db.Set("gorm:association_autoupdate", false).Model(eventItem).Omit("id", "account_id", "updated_at", "created_at").Update(input).Error; err != nil { return err}

	go EventListener{}.ReloadEventHandlers()

	return nil
}

func (eventItem EventItem) delete () error {
	return db.Model(EventItem{}).Where("id = ?", eventItem.ID).Delete(eventItem).Error
}