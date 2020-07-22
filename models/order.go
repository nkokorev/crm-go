package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type Order struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"accountId" gorm:"index,not null"` // аккаунт-владелец ключа

	// Описание заказа, может быть видно пользователю
	Description string 	`json:"description" gorm:"type:varchar(255);default:''"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ############# Entity interface #############
func (order Order) GetId() uint { return order.ID }
func (order *Order) setId(id uint) { order.ID = id }
func (order Order) GetAccountId() uint { return order.AccountID }
func (order *Order) setAccountId(id uint) { order.AccountID = id }
func (Order) systemEntity() bool { return false }

// ############# Entity interface #############

func (Order) PgSqlCreate() {
	db.CreateTable(&Order{})
	db.Model(&Order{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}
func (order *Order) BeforeCreate(scope *gorm.Scope) error {
	order.ID = 0
	return nil
}
func (Order) TableName() string {
	return "web_hooks"
}

// ######### CRUD Functions ############
func (order Order) create() (Entity, error)  {
	// if err := db.Create(&order).Find(&order, order.ID).Error; err != nil {
	wb := order
	if err := db.Create(&wb).Error; err != nil {
		return nil, err
	}

	var entity Entity = &wb

	return entity, nil
}

func (Order) get(id uint) (Entity, error) {

	var order Order

	err := db.First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}
func (order *Order) load() error {
	if order.ID < 1 {
		return utils.Error{Message: "Невозможно загрузить Order - не указан  ID"}
	}

	err := db.First(order).Error
	if err != nil {
		return err
	}
	return nil
}

func (Order) getList(accountId uint, sortBy string) ([]Entity, uint, error) {

	webHooks := make([]Order,0)
	var total uint

	err := db.Model(&Order{}).Limit(1000).Order(sortBy).Where( "account_id = ?", accountId).
		Find(&webHooks).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, 0, err
	}

	// Определяем total
	err = db.Model(&Order{}).Where("account_id = ?", accountId).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(webHooks))
	for i,_ := range webHooks {
		entities[i] = &webHooks[i]
	}

	return entities, total, nil
}

func (Order) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	webHooks := make([]Order,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&Order{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Order{}).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&Order{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Order{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(webHooks))
	for i,_ := range webHooks {
		entities[i] = &webHooks[i]
	}

	return entities, total, nil
}

func (Order) getByEvent(eventName string) (*Order, error) {

	wh := Order{}

	if err := db.First(&wh, "event_type = ?", eventName).Error; err != nil {
		return nil, err
	}

	return &wh, nil
}

func (order *Order) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).
		Model(order).Where("id", order.ID).Omit("id", "account_id").Updates(input).Error
}

func (order Order) delete () error {
	return db.Model(Order{}).Where("id = ?", order.ID).Delete(order).Error
}
// ######### END CRUD Functions ############


// ########## Work function ############

