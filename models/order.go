package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type Order struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"accountID" gorm:"index,not null"` // аккаунт-владелец ключа

	Description string 	`json:"description" gorm:"type:varchar(255);"`
	// Описание заказа, может быть видно пользователю
	Name		string 	`json:"name" gorm:"type:varchar(255);"` // Имя заказа

	////// Данные заказа ///////
	// Юр.лицо / Физ.лицо
	Type	string `json:"type"`

	// Магазин (сайт) с которого пришел заказ. НЕ может быть null.
	WebSiteID 	uint	`json:"webSiteID"`
	WebSite	WebSite

	// Способ (канал) заказа: "Заказ из корзины", "Заказ по телефону", "Пропущенный звонок", "Письмо.."
	OrderSourceID string	`json:"orderSourceID"`
	OrderSource	OrderSource `json:"orderSource"`
	
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}

// ############# Entity interface #############
func (order Order) GetID() uint { return order.ID }
func (order *Order) setID(id uint) { order.ID = id }
func (order Order) GetAccountID() uint { return order.AccountID }
func (order *Order) setAccountID(id uint) { order.AccountID = id }
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
	order.ID = 0
	return nil
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

	err := db.First(order, order.ID).Error
	if err != nil {
		return err
	}
	return nil
}

func (Order) getList(accountID uint, sortBy string) ([]Entity, uint, error) {

	orders := make([]Order,0)
	var total uint

	err := db.Model(&Order{}).Limit(100).Order(sortBy).Where( "account_id = ?", accountID).
		Find(&orders).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, 0, err
	}

	// Определяем total
	err = db.Model(&Order{}).Where("account_id = ?", accountID).Count(&total).Error
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

func (Order) getPaginationList(accountID uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	orders := make([]Order,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&Order{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountID).
			Find(&orders, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Order{}).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountID, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&Order{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountID).
			Find(&orders).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Order{}).Where("account_id = ?", accountID).Count(&total).Error
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

