package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type Order struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"accountId" gorm:"index,not null"` // аккаунт-владелец ключа

	Name string `json:"name" gorm:"type:varchar(255);default:'New api key';"` //
	Enabled bool `json:"enabled" gorm:"type:bool;default:true"` // активен ли ключ

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Order) PgSqlCreate() {
	db.CreateTable(&Order{})

	db.Model(&Order{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}

// ############# Entity interface #############
func (order Order) getId() uint           { return order.ID }
func (order *Order) setId(id uint)        { order.ID = id }
func (order Order) GetAccountId() uint    { return order.AccountID }
func (order *Order) setAccountId(id uint) { order.AccountID = id }
func (Order) systemEntity() bool { return false }
// ############# Entity interface #############


// ###### GORM Functional #######
func (Order) TableName() string { return "orders" }
func (order *Order) BeforeCreate(scope *gorm.Scope) error {
	order.ID = 0
	return nil
}
// ###### End of GORM Functional #######

// ############# CRUD Entity interface #############

func (order Order) create() (Entity, error)  {

	ord := order
	if err := db.Create(&ord).Error; err != nil {
		return nil, err
	}
	var entity Entity = &ord

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

	err := db.First(order).Error
	if err != nil {
		return err
	}
	return nil
}

func (Order) getList(accountId uint, sortBy string) ([]Entity, uint, error) {

	orders := make([]Order,0)
	var total uint

	// if need to search
	err := db.Model(&Order{}).Limit(1000).Order(sortBy).Where( "account_id = ?", accountId).
		Find(&orders).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, 0, err
	}

	// Определяем total
	err = db.Model(&Order{}).Where("account_id = ?", accountId).Count(&total).Error
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

		err := db.Model(&Order{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&orders, "name ILIKE ? OR code ILIKE ? OR postal_code_from ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Order{}).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR postal_code_from ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&Order{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
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
	return db.Set("gorm:association_autoupdate", false).Model(order).Omit("id", "account_id").Updates(input).Error
}

func (order Order) delete () error {
	return db.Model(Order{}).Where("id = ?", order.ID).Delete(order).Error
}

// ########## End of CRUD Entity interface ###########