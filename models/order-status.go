package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"log"
	"time"
)

type OrderStatus struct {
	Id     			uint   	`json:"id" gorm:"primary_key"`
	AccountId 		uint 	`json:"-" gorm:"type:int;index;not null;"`

	// new, canceled, ...
	Code	string 	`json:"code" gorm:"type:varchar(32);"`

	// new, agreement, equipment, delivery, completed, canceled
	Group	string 	`json:"group" gorm:"type:varchar(32);"`
	GroupName	string 	`json:"group" gorm:"type:varchar(128);"`

	// 'Новый заказ', 'Передан в комплектацию', 'Отменен'
	Name	string `json:"name" gorm:"type:varchar(128);"`

	Description string 	`json:"description" gorm:"type:varchar(255);"` // Описание назначения канала

	CreatedAt 		time.Time `json:"createdAt"`
	UpdatedAt 		time.Time `json:"updatedAt"`
}

// ############# Entity interface #############
func (orderStatus OrderStatus) GetId() uint { return orderStatus.Id }
func (orderStatus *OrderStatus) setId(id uint) { orderStatus.Id = id }
func (orderStatus *OrderStatus) setPublicId(id uint) { orderStatus.Id = id }
func (orderStatus OrderStatus) GetAccountId() uint { return orderStatus.AccountId }
func (orderStatus *OrderStatus) setAccountId(id uint) { orderStatus.AccountId = id }
func (orderStatus OrderStatus) SystemEntity() bool { return orderStatus.AccountId == 1 }

// ############# Entity interface #############

func (OrderStatus) PgSqlCreate() {
	db.AutoMigrate(&OrderStatus{})
	db.Model(&OrderStatus{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")

	mainAccount, err := GetMainAccount()
	if err != nil {
		log.Println("Не удалось найти главный аккаунт для OrderStatus")
	}

	orderStatuses := []OrderStatus{
		// new, agreement, equipment, delivery, completed, canceled
		{Name: "Новый заказ", 	Code: "new", Group:"new", GroupName:"Необработанные заявки",	Description: "Необработанный заказ, первоначальный статус заказа."},
		{Name: "Заказ подтвержден", 	Code: "agreement_order", Group: "agreement", 	GroupName:"Согласование", Description: "Заказ подтвержден."},
		{Name: "Предложена замена", 	Code: "agreement_change", Group: "agreement", 	GroupName:"Согласование", Description: "Предложена замена."},
		{Name: "Согласование с клиентом", 	Code: "agreement_approval", Group: "agreement", GroupName:"Согласование", Description: "Согласование с клиентом."},
		{Name: "Комплектуется", 	Code: "equipping", Group: "equipment", 		GroupName:"Комплектация", Description: "Комплектуется."},
		{Name: "Укомплектован", 	Code: "equipped", Group: "equipment", 		GroupName:"Комплектация", Description: "Укомплектован"},
		{Name: "Передан в доставку", 	Code: "delivery_sent", 	Group: "delivery", 	GroupName:"Доставка", Description: "Заказ передан в доставку"},
		{Name: "Доставляется", 			Code: "delivering", 	Group: "delivery", 				GroupName:"Доставка", Description: "Заказ в процессе доставки"},
		{Name: "Доставка перенесена", 	Code: "delivery_rescheduled", Group: "delivery", GroupName:"Доставка", Description: "Доставка перенесена"},
		{Name: "Заказ выполнен", 	Code: "completed", Group: "completed", GroupName:"Выполнение", Description: "Заказ выполнен (завершен)"},

		{Name: "Недозвон", 			Code: "canceled_call", 			Group: "canceled", GroupName:"Отмена", Description: "Заказ отменен"},
		{Name: "Нет в наличии", 	Code: "canceled_out_of_stock", 	Group: "canceled", GroupName:"Отмена", Description: "Заказ отменен"},
		{Name: "Не устроила цена", 	Code: "canceled_price", 		Group: "canceled", GroupName:"Отмена", Description: "Заказ отменен"},
		{Name: "Не устроила доставка", 	Code: "canceled_delivery", 	Group: "canceled", GroupName:"Отмена", Description: "Заказ отменен"},
		{Name: "Отменен", 			Code: "canceled_any", 			Group: "canceled", GroupName:"Отмена", Description: "Заказ отменен"},
	}
	for _,v := range orderStatuses {
		_, err = mainAccount.CreateEntity(&v)
		if err != nil {
			log.Fatal(err)
		}
	}
}
func (orderStatus *OrderStatus) BeforeCreate(scope *gorm.Scope) error {
	orderStatus.Id = 0
	return nil
}


// ######### CRUD Functions ############
func (orderStatus OrderStatus) create() (Entity, error)  {

	_orderStatus := orderStatus

	if err := db.Create(&_orderStatus).Error; err != nil {
		return nil, err
	}

	var newItem Entity = &_orderStatus

	return newItem, nil
}

func (OrderStatus) get(id uint) (Entity, error) {

	var orderStatus OrderStatus

	err := db.First(&orderStatus, id).Error
	if err != nil {
		return nil, err
	}
	return &orderStatus, nil
}
func (orderStatus *OrderStatus) load() error {

	err := db.First(orderStatus, orderStatus.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (*OrderStatus) loadByPublicId() error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}

func (OrderStatus) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return OrderStatus{}.getPaginationList(accountId, 0, 100, sortBy, "")
}
func (OrderStatus) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	orderStatuses := make([]OrderStatus,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := db.Model(&OrderStatus{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id IN (?)", []uint{1, accountId}).
			Find(&orderStatuses, "name ILIKE ? OR description ILIKE ?", search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&OrderStatus{}).
			Where("account_id = ? AND name ILIKE ? OR description ILIKE ? ", accountId, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&OrderStatus{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id IN (?)", []uint{1, accountId}).
			Find(&orderStatuses).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&OrderStatus{}).Where("account_id IN (?)", []uint{1, accountId}).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(orderStatuses))
	for i,_ := range orderStatuses {
		entities[i] = &orderStatuses[i]
	}

	return entities, total, nil
}

func (orderStatus *OrderStatus) update(input map[string]interface{}) error {

	// work!!!
	if err := db.Set("gorm:association_autoupdate", false).Model(orderStatus).Omit("id", "account_id","created_at").
		Updates(input).Error; err != nil {
		return err
	}

	err := db.First(orderStatus, orderStatus.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (orderStatus *OrderStatus) delete () error {
	return db.Model(OrderStatus{}).Where("id = ?", orderStatus.Id).Delete(orderStatus).Error
}
// ######### END CRUD Functions ############
