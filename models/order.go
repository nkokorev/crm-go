package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type orderType = string

type Order struct {
	Id     		uint   	`json:"id" gorm:"primary_key"`
	AccountId 	uint 	`json:"accountId" gorm:"index,not null"` // аккаунт-владелец ключа

	// Комментарий клиента к заказу
	CustomerComments string	`json:"description" gorm:"type:varchar(255);"`

	// Комментарии к заказу
	Comments	[]OrderComment `json:"comments"`

	////// Данные заказа ///////
	Individual	bool `json:"individual" gorm:"type:bool;default:true;not null;"` // Физ.лицо - true, Юрлицо - false

	// Магазин (сайт) с которого пришел заказ. НЕ может быть null.
	WebSiteId 	uint	`json:"webSiteId" gorm:"type:int;not null;"`
	WebSite		WebSite	`json:"webSite"`

	// Способ (канал) заказа: "Заказ из корзины", "Заказ по телефону", "Пропущенный звонок", "Письмо.."
	OrderChannelId 	uint	`json:"orderChannelId" gorm:"type:int;not null;"`
	OrderChannel 	OrderChannel `json:"orderChannel"`

	CreatedAt time.Time 	`json:"createdAt"`
	UpdatedAt time.Time 	`json:"updatedAt"`
	DeletedAt *time.Time 	`json:"deletedAt"`
}

// ############# Entity interface #############
func (order Order) GetId() uint { return order.Id }
func (order *Order) setId(id uint) { order.Id = id }
func (order Order) GetAccountId() uint { return order.AccountId }
func (order *Order) setAccountId(id uint) { order.AccountId = id }
func (Order) SystemEntity() bool { return false }

// ############# Entity interface #############

func (Order) PgSqlCreate() {
	if !db.HasTable(&Order{}) {
		db.CreateTable(&Order{})
	}
	db.Model(&Order{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	
	// fmt.Println("Щквук!")
}
func (order *Order) BeforeCreate(scope *gorm.Scope) error {
	order.Id = 0
	return nil
}

// ######### CRUD Functions ############
func (order Order) create() (Entity, error)  {
	// if err := db.Create(&order).Find(&order, order.Id).Error; err != nil {
	wb := order
	if err := db.Create(&wb).Error; err != nil {
		return nil, err
	}

	var entity Entity = &wb

	return entity, nil
}
func (Order) get(id uint) (Entity, error) {

	var order Order

	err := db.Preload("WebSite").Preload("OrderChannel").First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}
func (order *Order) load() error {

	if order.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить Order - не указан  Id"}
	}

	err := db.Preload("WebSite").Preload("OrderChannel").First(order, order.Id).Error
	if err != nil {
		fmt.Println("Ja!")
		return err
	}

	return nil
}
func (Order) getList(accountId uint, sortBy string) ([]Entity, uint, error) {

	orders := make([]Order,0)
	var total uint

	err := db.Preload("WebSite").Preload("OrderChannel").Model(&Order{}).Limit(100).Order(sortBy).Where( "account_id = ?", accountId).
		Find(&orders).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, 0, err
	}

	// Определяем total
	err = db.Preload("WebSite").Preload("OrderChannel").Model(&Order{}).Where("account_id = ?", accountId).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(orders))
	for i,_ := range orders {
		entities[i] = &orders[i]
	}

	return entities, total, nil
}
func (Order) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	orders := make([]Order,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&Order{}).Preload("WebSite").Preload("OrderChannel").Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&orders, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
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

		err := db.Model(&Order{}).Limit(limit).Preload("WebSite").Preload("OrderChannel").Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&orders).Error
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
	entities := make([]Entity,len(orders))
	for i,_ := range orders {
		entities[i] = &orders[i]
	}

	return entities, total, nil
}

func (order *Order) update(input map[string]interface{}) error {

	delete(input,"webSite")
	delete(input,"orderChannel")

	return db.Set("gorm:association_autoupdate", false).
		Model(order).Preload("WebSite").Preload("OrderChannel").Omit("id", "account_id").Updates(input).Error
}

func (order Order) delete () error {
	return db.Model(Order{}).Where("id = ?", order.Id).Delete(order).Error
}
// ######### END CRUD Functions ############


// ########## Work function ############

