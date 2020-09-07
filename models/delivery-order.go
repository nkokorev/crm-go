package models

import (
	"database/sql"
	"fmt"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)
type DeliveryOrder struct {
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	PublicId	uint   	`json:"public_id" gorm:"type:int;index;not null;"` // Публичный ID заказа внутри магазина
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	// Данные заказа
	OrderId 	*uint	`json:"order_id" gorm:"type:int;"`
	// Order	*Order	`json:"-"`

	// Данные заказчика
	CustomerId 	uint	`json:"customer_id" gorm:"type:int;not null"`
	Customer	User	`json:"customer"`

	// Магазин (сайт) с которого пришел заказ. НЕ может быть null.
	WebSiteId 	uint	`json:"web_site_id" gorm:"type:int;not null;"`
	WebSite		WebSite	`json:"web_site"`

	// Тип доставки
	Code		string 	`json:"code" gorm:"type:varchar(32);"`
	MethodId 	uint	`json:"method_id" gorm:"type:int;not null;"`

	Address		string 	`json:"address" gorm:"type:varchar(255);"`
	PostalCode	string 	`json:"postal_code" gorm:"type:varchar(32);"`
	Delivery	Delivery	`json:"delivery" gorm:"-"` // << preload

	// Фиксируем стоимость
	AmountId  	uint			`json:"amount_id" gorm:"type:int;not null;"`
	Amount  	PaymentAmount	`json:"amount"`

	// Статус заказа
	StatusId  	uint			`json:"status_id" gorm:"type:int;"`
	Status		DeliveryStatus	`json:"status" gorm:"preload"`

	CreatedAt 	time.Time `json:"created_at"`
	UpdatedAt 	time.Time `json:"updated_at"`
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
	if err := db.Migrator().CreateTable(&DeliveryOrder{}); err != nil {log.Fatal(err)}
	// db.Model(&DeliveryOrder{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&DeliveryOrder{}).AddForeignKey("status_id", "delivery_statuses(id)", "RESTRICT", "CASCADE")
	err := db.Exec("ALTER TABLE delivery_orders " +
		"ADD CONSTRAINT delivery_orders_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT delivery_orders_status_id_fkey FOREIGN KEY (status_id) REFERENCES delivery_statuses(id) ON DELETE RESTRICT ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (deliveryOrder *DeliveryOrder) BeforeCreate(tx *gorm.DB) error {
	deliveryOrder.Id = 0

	var lastIdx sql.NullInt64
	err := db.Model(&DeliveryOrder{}).Where("account_id = ?",  deliveryOrder.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }

	deliveryOrder.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (deliveryOrder *DeliveryOrder) AfterFind(tx *gorm.DB) (err error) {

	delivery, err := Account{Id: deliveryOrder.AccountId}.GetDeliveryByCode(deliveryOrder.Code, deliveryOrder.MethodId)
	if err == nil {
		deliveryOrder.Delivery = delivery
		return
	}
	return nil
}
func (deliveryOrder *DeliveryOrder) AfterCreate(tx *gorm.DB) error {
	event.AsyncFire(Event{}.DeliveryOrderCreated(deliveryOrder.AccountId, deliveryOrder.Id))
	return nil
}
func (deliveryOrder *DeliveryOrder) AfterUpdate(tx *gorm.DB) (err error) {

	event.AsyncFire(Event{}.DeliveryOrderUpdated(deliveryOrder.AccountId, deliveryOrder.Id))

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

	_item := deliveryOrder
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}

func (DeliveryOrder) get(id uint, preloads []string) (Entity, error) {

	var item DeliveryOrder

	err := item.GetPreloadDb(false, false, preloads).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
func (deliveryOrder *DeliveryOrder) load(preloads []string) error {
	if deliveryOrder.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить DeliveryOrder - не указан  Id"}
	}

	err := deliveryOrder.GetPreloadDb(false, false, preloads).First(deliveryOrder, deliveryOrder.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (deliveryOrder *DeliveryOrder) loadByPublicId(preloads []string) error {

	if deliveryOrder.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить DeliveryOrder - не указан  Id"}
	}
	if err := deliveryOrder.GetPreloadDb(false,false, preloads).First(deliveryOrder, "account_id = ? AND public_id = ?", deliveryOrder.AccountId, deliveryOrder.PublicId).Error; err != nil {
		return err
	}
	return nil
}

func (DeliveryOrder) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return DeliveryOrder{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (DeliveryOrder) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {
	
	deliveryOrders := make([]DeliveryOrder,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		// string pattern
		// jsearch := search
		search = "%"+search+"%"

		err := (&DeliveryOrder{}).GetPreloadDb(false,false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&deliveryOrders, "address ILIKE ? OR postal_code ILIKE ?", search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&DeliveryOrder{}).GetPreloadDb(false,false, nil).
			Where("account_id = ? AND address ILIKE ? OR postal_code ILIKE ? ", accountId, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&DeliveryOrder{}).GetPreloadDb(false,false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&deliveryOrders).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&DeliveryOrder{}).GetPreloadDb(false,false, nil).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(deliveryOrders))
	for i := range deliveryOrders {
		entities[i] = &deliveryOrders[i]
	}

	return entities, total, nil
}

func (deliveryOrder *DeliveryOrder) update(input map[string]interface{}, preloads []string) error {

	delete(input, "customer")
	delete(input, "web_site")
	delete(input, "delivery")
	delete(input, "amount")
	delete(input, "status")
	
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id","order_id","customer_id","web_site_id","method_id","amount_id","status_id"}); err != nil {
		return err
	}

	// Отметка на будущее для события
	_newStatusId, ok := input["statusId"].(float64)
	_oldStatusId :=  deliveryOrder.StatusId

	// if err := db.Set("gorm:association_autoupdate", false).Model(deliveryOrder).Where("id = ?", deliveryOrder.Id).
	// if err := db.Model(&DeliveryOrder{}).Where("id = ?", deliveryOrder.Id).
	if err := deliveryOrder.GetPreloadDb(false,false, nil).Where("id = ?", deliveryOrder.Id).
		Omit("id", "account_id","created_at","public_id").Updates(input).Error; err != nil {
		return err
	}

	err := deliveryOrder.GetPreloadDb(false,false, preloads).First(deliveryOrder, deliveryOrder.Id).Error
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
			err := Account{Id: deliveryOrder.AccountId}.LoadEntity(&order, *deliveryOrder.OrderId, nil)
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
				err := Account{Id: deliveryOrder.AccountId}.LoadEntity(&payment, order.Payment.Id, nil)
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
					if err := order.update(map[string]interface{}{"statusId":status.Id},nil); err != nil {
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

	var idx = make([]uint,0)
	idx = append(idx,deliveryOrder.AmountId)

	if err := (PaymentAmount{}).deletes(idx); err != nil {
		return err
	}

	
	return deliveryOrder.GetPreloadDb(true,false,nil).Where("id = ?", deliveryOrder.Id).Delete(deliveryOrder).Error
}
// ######### END CRUD Functions ############

func (deliveryOrder *DeliveryOrder) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(&deliveryOrder)
	} else {
		_db = _db.Model(&DeliveryOrder{})
	}

	if autoPreload {
		return db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"Amount","WebSite","Customer","Status"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}

