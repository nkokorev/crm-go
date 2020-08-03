package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"log"
	"time"
)

type DeliveryOrder struct {
	Id     		uint   	`json:"id" gorm:"primary_key"`
	PublicId	uint   	`json:"publicId" gorm:"type:int;index;not null;"` // Публичный ID заказа внутри магазина
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	// Данные заказа
	OrderId 	uint	`json:"orderId" gorm:"type:int;not null"`
	// Order	*Order	`json:"-"`

	// Данные заказчика
	CustomerId 	uint	`json:"customerId" gorm:"type:int;not null"`
	Customer	User	`json:"customer"`

	// Магазин (сайт) с которого пришел заказ. НЕ может быть null.
	WebSiteId 	uint	`json:"webSiteId" gorm:"type:int;not null;"`
	WebSite		WebSite	`json:"webSite"`

	// Тип доставки
	Code		string 	`json:"deliveryCode" gorm:"type:varchar(32);"`
	MethodId 	uint	`json:"methodId" gorm:"type:int;not null;"`

	Address		string 	`json:"address" gorm:"type:varchar(255);"`
	PostalCode	string 	`json:"postalCode" gorm:"type:varchar(32);"`
	Delivery	Delivery	`json:"delivery" gorm:"-"` // << preload

	// Фиксируем стоимость
	AmountId  	uint			`json:"amountId" gorm:"type:int;not null;"`
	Amount  	PaymentAmount	`json:"amount"`

	// Статус заказа
	StatusId  	uint			`json:"statusId" gorm:"type:int;"`
	Status		DeliveryStatus	`json:"status" gorm:"preload"`

	CreatedAt 	time.Time `json:"createdAt"`
	UpdatedAt 	time.Time `json:"updatedAt"`
}

// ############# Entity interface #############
func (deliveryOrder DeliveryOrder) GetId() uint { return deliveryOrder.Id }
func (deliveryOrder *DeliveryOrder) setId(id uint) { deliveryOrder.Id = id }
func (deliveryOrder *DeliveryOrder) setPublicId(publicId uint) { deliveryOrder.PublicId = publicId }
func (deliveryOrder DeliveryOrder) GetAccountId() uint { return deliveryOrder.AccountId }
func (deliveryOrder *DeliveryOrder) setAccountId(id uint) { deliveryOrder.AccountId = id }
func (DeliveryOrder) SystemEntity() bool { return false }

// ############# Entity interface #############

func (DeliveryOrder) PgSqlCreate() {
	db.AutoMigrate(&DeliveryOrder{})
	db.Model(&DeliveryOrder{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&DeliveryOrder{}).AddForeignKey("order_id", "orders(id)", "CASCADE", "CASCADE")
	db.Model(&DeliveryOrder{}).AddForeignKey("status_id", "delivery_statuses(id)", "CASCADE", "CASCADE")
}
func (deliveryOrder *DeliveryOrder) BeforeCreate(scope *gorm.Scope) error {
	deliveryOrder.Id = 0

	lastIdx := uint(0)
	var ord DeliveryOrder

	err := db.Where("account_id = ?", deliveryOrder.AccountId).Select("public_id").Last(&ord).Error;
	if err != nil && err != gorm.ErrRecordNotFound { return err}
	if err == gorm.ErrRecordNotFound {
		lastIdx = 0
	} else {
		lastIdx = ord.PublicId
	}
	deliveryOrder.PublicId = lastIdx + 1

	return nil
}
func (deliveryOrder *DeliveryOrder) AfterFind() (err error) {

	delivery, err := Account{Id: deliveryOrder.AccountId}.GetDeliveryByCode(deliveryOrder.Code, deliveryOrder.MethodId)
	if err == nil {
		deliveryOrder.Delivery = delivery
		return
	}
	return nil
}
func (deliveryOrder *DeliveryOrder) AfterCreate(scope *gorm.Scope) (error) {
	event.AsyncFire(Event{}.DeliveryOrderCreated(deliveryOrder.AccountId, deliveryOrder.PublicId))
	return nil
}
func (deliveryOrder *DeliveryOrder) AfterUpdate(tx *gorm.DB) (err error) {

	event.AsyncFire(Event{}.DeliveryOrderUpdated(deliveryOrder.AccountId, deliveryOrder.PublicId))

/*	orderStatusEntity, err := DeliveryStatus{}.get(deliveryOrder.StatusId)
	if err == nil && orderStatusEntity.GetAccountId() == deliveryOrder.AccountId {
		if deliveryStatus, ok := orderStatusEntity.(*DeliveryStatus); ok {
			if deliveryStatus.Code == "completed" {
				event.AsyncFire(Event{}.DeliveryOrderCompleted(deliveryOrder.AccountId, deliveryOrder.Id))
			}
			if deliveryStatus.Code == "canceled" {
				event.AsyncFire(Event{}.DeliveryOrderCanceled(deliveryOrder.AccountId, deliveryOrder.Id))
			}
		}

	}*/

	return nil
}
func (deliveryOrder *DeliveryOrder) AfterDelete(tx *gorm.DB) (err error) {
	event.AsyncFire(Event{}.DeliveryOrderDeleted(deliveryOrder.AccountId, deliveryOrder.Id))
	return nil
}

// ######### CRUD Functions ############
func (deliveryOrder DeliveryOrder) create() (Entity, error)  {

	_deliveryOrder := deliveryOrder

	if err := db.Create(&_deliveryOrder).Error; err != nil {
		return nil, err
	}

	var newItem Entity = &_deliveryOrder

	return newItem, nil
}

func (DeliveryOrder) get(id uint) (Entity, error) {

	var deliveryOrder DeliveryOrder

	err := db.Preload("WebSite").Preload("Amount").Preload("Order").Preload("Customer").First(&deliveryOrder, id).Error
	if err != nil {
		return nil, err
	}
	return &deliveryOrder, nil
}
func (deliveryOrder *DeliveryOrder) load() error {

	err := deliveryOrder.GetPreloadDb(false,false, true).First(deliveryOrder, deliveryOrder.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (deliveryOrder *DeliveryOrder) loadByPublicId() error {

	if deliveryOrder.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить DeliveryOrder - не указан  Id"}
	}
	if err := deliveryOrder.GetPreloadDb(false,false, true).First(deliveryOrder, "public_id = ?", deliveryOrder.PublicId).Error; err != nil {
		return err
	}
	return nil
}

func (DeliveryOrder) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return DeliveryOrder{}.getPaginationList(accountId, 0, 100, sortBy, "")
}
func (DeliveryOrder) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {
	
	deliveryOrders := make([]DeliveryOrder,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		// jsearch := search
		search = "%"+search+"%"

		err := (&DeliveryOrder{}).GetPreloadDb(false,false, true).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&deliveryOrders, "address ILIKE ? OR postal_code ILIKE ?", search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&DeliveryOrder{}).
			Where("account_id = ? AND address ILIKE ? OR postal_code ILIKE ? ", accountId, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&DeliveryOrder{}).GetPreloadDb(false,false, true ).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&deliveryOrders).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&DeliveryOrder{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(deliveryOrders))
	for i,_ := range deliveryOrders {
		entities[i] = &deliveryOrders[i]
	}

	return entities, total, nil
}

func (deliveryOrder *DeliveryOrder) update(input map[string]interface{}) error {

	delete(input, "order")
	delete(input, "customer")
	delete(input, "Customer")


	// Отметка на будущее для события
	_newStatusId, ok := input["statusId"].(float64)
	_oldStatusId :=  deliveryOrder.StatusId

	// if err := db.Set("gorm:association_autoupdate", false).Model(deliveryOrder).Where("id = ?", deliveryOrder.Id).
	// if err := db.Model(&DeliveryOrder{}).Where("id = ?", deliveryOrder.Id).
	if err := deliveryOrder.GetPreloadDb(true,false, false).Where("id = ?", deliveryOrder.Id).
		Omit("id", "account_id","created_at").Update(input).Error; err != nil {
		return err
	}

	err := deliveryOrder.GetPreloadDb(true,false, true).First(deliveryOrder, deliveryOrder.Id).Error
	if err != nil {
		return err
	}

	// Если флаг статуса доставки был изменен
	if ok && (_newStatusId != float64(_oldStatusId)) {

		switch deliveryOrder.Status.Group {
		case "agreement":
			event.AsyncFire(Event{}.DeliveryOrderAlimented(deliveryOrder.AccountId, deliveryOrder.PublicId))
		case "delivery":
			event.AsyncFire(Event{}.DeliveryOrderInProcess(deliveryOrder.AccountId, deliveryOrder.PublicId))
		case "completed":
			// Обновляем платеж
			var order Order
			err := Account{Id: deliveryOrder.AccountId}.LoadEntity(&order, deliveryOrder.OrderId)
			if err != nil {
				return utils.Error{Message: "Не удалось обновить статус платежа - не найден заказ"}
			}

			paymentMethod, err := Account{Id: deliveryOrder.AccountId}.GetPaymentMethodByType(order.PaymentMethodType, order.PaymentMethodId)
			if err != nil {
				return utils.Error{Message: "Не удалось обновить статус платежа - не найден заказ"}
			}

			// Если тип оплаты яндекс и платеж разнесен во времени
			if paymentMethod.GetCode() == "payment_yandex" && !paymentMethod.IsInstantDelivery() {
				// 1. Надо узнать External-ID чека
				var payment Payment
				err := Account{Id: deliveryOrder.AccountId}.LoadEntity(&payment, order.Payment.Id)
				if err != nil {
					return utils.Error{Message: "Не удалось обновить статус платежа - не найден платеж"}
				}

				if payment.ExternalId != "" {
					_, err = paymentMethod.PrepaymentCheck(&payment, order)
					if err != nil {
						fmt.Println("Error: ", err)
					}
				}

				status, err := OrderStatus{}.GetCompletedStatus()
				if err == nil {
					if err := order.update(map[string]interface{}{"statusId":status.Id}); err != nil {
						log.Println(err)
					}
				}
			}

			// todo: !!! Перевести ЗАКАЗ в статус - Выполнено!!

			event.AsyncFire(Event{}.DeliveryOrderCompleted(deliveryOrder.AccountId, deliveryOrder.PublicId))
		case "canceled":
			event.AsyncFire(Event{}.DeliveryOrderCanceled(deliveryOrder.AccountId, deliveryOrder.PublicId))
		}

		// общая информация об обновлении статуса
		event.AsyncFire(Event{}.DeliveryOrderStatusUpdated(deliveryOrder.AccountId, deliveryOrder.PublicId))


	}

	// see: AfterUpdated()

	return nil
}

func (deliveryOrder *DeliveryOrder) delete () error {
	return db.Model(DeliveryOrder{}).Where("id = ?", deliveryOrder.Id).Delete(deliveryOrder).Error
}
// ######### END CRUD Functions ############

func (deliveryOrder *DeliveryOrder) GetPreloadDb(autoUpdate bool, getModel bool, preload bool) *gorm.DB {
	_db := db

	if autoUpdate {
		_db = _db.Set("gorm:association_autoupdate", false)
	}
	if getModel {
		_db = _db.Model(&deliveryOrder)
	} else {
		_db = _db.Model(&DeliveryOrder{})
	}

	if preload {
		return _db.Preload("WebSite").Preload("Amount").Preload("Customer").Preload("Status")
	} else {
		return _db
	}
	

}

