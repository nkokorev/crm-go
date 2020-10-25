package models

import (
	"database/sql"
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
	Order		*Order	`json:"order"`

	// Данные заказчика
	CustomerId 	uint	`json:"customer_id" gorm:"type:int;not null"`
	Customer	User	`json:"customer"`

	// Магазин (сайт) с которого пришел заказ. НЕ может быть null.
	WebSiteId 	uint	`json:"web_site_id" gorm:"type:int;not null;"`
	WebSite		WebSite	`json:"web_site"`

	// Тип доставки
	Code		string 	`json:"code" gorm:"type:varchar(32);"`
	MethodId 	uint	`json:"method_id" gorm:"type:int;not null;"`

	// Адрес по карте
	Address		*string 	`json:"address" gorm:"type:varchar(255);"` // << null if pickup
	PostalCode	*string 	`json:"postal_code" gorm:"type:varchar(32);"` // << null if pickup
	Delivery	Delivery	`json:"delivery" gorm:"-"` // << preload

	// Комментарий клиента по доставке ()
	Comment 	*string		`json:"comment" gorm:"type:varchar(255);"`

	// Стоимость доставки
	Cost		float64 	`json:"cost" gorm:"type:numeric;default:0"`

	// Расчетный вес заказа
	Weight 		*float64		`json:"weight" gorm:"type:numeric;default:0"`

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
	// AsyncFire(*Event{}.DeliveryOrderCreated(deliveryOrder.AccountId, deliveryOrder.Id))
	AsyncFire(NewEvent("DeliveryOrderCreated", map[string]interface{}{"account_id":deliveryOrder.AccountId, "delivery_order_id":deliveryOrder.Id}))
	return nil
}
func (deliveryOrder *DeliveryOrder) AfterUpdate(tx *gorm.DB) (err error) {

	// AsyncFire(*Event{}.DeliveryOrderUpdated(deliveryOrder.AccountId, deliveryOrder.Id))
	AsyncFire(NewEvent("DeliveryOrderUpdated", map[string]interface{}{"account_id":deliveryOrder.AccountId, "delivery_order_id":deliveryOrder.Id}))

/*	orderStatusEntity, err := DeliveryStatus{}.get(deliveryOrder.StatusId)
	if err == nil && orderStatusEntity.GetAccountId() == deliveryOrder.AccountId {
		if deliveryStatus, ok := orderStatusEntity.(*DeliveryStatus); ok {
			if deliveryStatus.Code == "completed" {
				AsyncFire(*Event{}.DeliveryOrderCompleted(deliveryOrder.AccountId, deliveryOrder.Id))
			}
			if deliveryStatus.Code == "canceled" {
				AsyncFire(*Event{}.DeliveryOrderCanceled(deliveryOrder.AccountId, deliveryOrder.Id))
			}
		}

	}*/

	return nil
}
func (deliveryOrder *DeliveryOrder) AfterDelete(tx *gorm.DB) (err error) {
	// AsyncFire(*Event{}.DeliveryOrderDeleted(deliveryOrder.AccountId, deliveryOrder.Id))
	AsyncFire(NewEvent("DeliveryOrderDeleted", map[string]interface{}{"account_id":deliveryOrder.AccountId, "delivery_order_id":deliveryOrder.Id}))
	return nil
}

// ######### CRUD Functions ############
func (deliveryOrder DeliveryOrder) create() (Entity, error)  {

	_item := deliveryOrder
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).Error; err != nil {
		return nil, err
	}
	_ = _item.load([]string{"Status"})

	// Добавляем в корзину стоимость доставки

	_item.createDeliveryItemIfNotExist()

	/*if _item.OrderId != nil {
		order := Order{Id: *_item.OrderId}
		if err := order.load([]string{"CartItems"}); err == nil {
			_ = order.AppendDeliveryItem(_item.Delivery, _item.Cost)
		}
	}*/

	var entity Entity = &_item

	return entity, nil
}

func (deliveryOrder *DeliveryOrder) createDeliveryItemIfNotExist(){
	if deliveryOrder.OrderId != nil {
		order := Order{Id: *deliveryOrder.OrderId}
		if err := order.load([]string{"CartItems"}); err == nil {
			if !order.ExistDeliveryItem() {
				_ = order.AppendDeliveryItem(deliveryOrder.Delivery, deliveryOrder.Cost)
			}
		}
	}
}
func (deliveryOrder *DeliveryOrder) updateDeliveryItem(input map[string]interface{}){
	if deliveryOrder.OrderId != nil {
		order := Order{Id: *deliveryOrder.OrderId}
		if err := order.load([]string{"CartItems"}); err == nil {
			_ = order.UpdateDeliveryItem(input)
		}
	}
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
	delete(input, "order")
	if err := deliveryOrder.load([]string{"Status"}); err != nil {
		return err
	}
	
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id","order_id","customer_id","web_site_id","method_id","status_id"}); err != nil {
		return err
	}

	// Отметка на будущее для события
	_newStatusId, ok := input["status_id"].(uint)
	_oldStatusId :=  deliveryOrder.StatusId

	// Создаем доставку, если ее нет
	deliveryOrder.createDeliveryItemIfNotExist()
	if costI, ok := input["cost"]; ok {
		cost, ok := costI.(float64)
		if ok {
			deliveryOrder.updateDeliveryItem(map[string]interface{}{"cost":cost})
		}
	}


	if err := deliveryOrder.GetPreloadDb(false,false, nil).Where("id = ?", deliveryOrder.Id).
		Omit("id", "account_id","created_at","public_id").Updates(input).Error; err != nil {
		return err
	}

	err := deliveryOrder.GetPreloadDb(false,false, preloads).First(deliveryOrder, deliveryOrder.Id).Error
	if err != nil {
		return err
	}

	// Если флаг статуса доставки был изменен
	if ok && (_newStatusId != uint(_oldStatusId)) {

		switch deliveryOrder.Status.Group {
		case "agreement":
			AsyncFire(NewEvent("DeliveryOrderAlimented", map[string]interface{}{"account_id":deliveryOrder.AccountId, "delivery_order_id":deliveryOrder.Id}))

		case "delivery":
			AsyncFire(NewEvent("DeliveryOrderInProcess", map[string]interface{}{"account_id":deliveryOrder.AccountId, "delivery_order_id":deliveryOrder.Id}))
		case "completed":
			// Обновляем платеж
			var order Order
			err := Account{Id: deliveryOrder.AccountId}.LoadEntity(&order, *deliveryOrder.OrderId, nil)
			if err != nil {
				// return utils.Error{Message: "Не удалось обновить статус платежа - не найден заказ"}
				log.Println("Не удалось обновить статус платежа - не найден заказ")
			} else {
				paymentMethod, err := Account{Id: deliveryOrder.AccountId}.GetPaymentMethodByType(order.PaymentMethodType, order.PaymentMethodId)
				if err != nil {
					log.Println("GetPaymentMethodByType: Не удалось обновить статус платежа - не найден заказ: ", err)
				} else {
					// Если тип оплаты яндекс и платеж разнесен во времени
					if paymentMethod.GetCode() == "payment_yandex" && !paymentMethod.IsInstantDelivery() {
						// 1. Надо узнать External-ID чека
						var payment Payment
						err := Account{Id: deliveryOrder.AccountId}.LoadEntity(&payment, order.Payment.Id, nil)
						if err != nil {
							log.Println("Не удалось обновить статус платежа - не найден платеж")
						} else {
							if payment.ExternalId != "" {
								_, err = paymentMethod.PrepaymentCheck(&payment, order)
								if err != nil {
									log.Println("Error ExternalId: ", err)
								}
							}
						}

						status, err := OrderStatus{}.GetCompletedStatus()
						if err == nil {
							if err := order.update(map[string]interface{}{"status_id":status.Id},nil); err != nil {
								log.Println(err)
							}
						}
					}
				}

				/* Изменяем статус заказа */
				if err := order.SetCompleted(); err != nil {log.Println("set completed order: ", err)}


			}

			AsyncFire(NewEvent("DeliveryOrderCompleted", map[string]interface{}{"account_id":deliveryOrder.AccountId, "delivery_order_id":deliveryOrder.Id}))
		case "canceled":


			var order Order
			err := Account{Id: deliveryOrder.AccountId}.LoadEntity(&order, *deliveryOrder.OrderId, nil)
			if err != nil {
				return utils.Error{Message: "Не удалось обновить статус платежа - не найден заказ"}
			} else {
				// Не факт что нужно отменять заказ
				// if err := order.SetCanceled(); err != nil {log.Println("SetCanceled order: ", err)}
			}

			AsyncFire(NewEvent("DeliveryOrderCanceled", map[string]interface{}{"account_id":deliveryOrder.AccountId, "delivery_order_id":deliveryOrder.Id}))

		}

		// общая информация об обновлении статуса
		// AsyncFire(*Event{}.DeliveryOrderStatusUpdated(deliveryOrder.AccountId, deliveryOrder.PublicId))
		AsyncFire(NewEvent("DeliveryOrderStatusUpdated", map[string]interface{}{"account_id":deliveryOrder.AccountId, "delivery_order_id":deliveryOrder.Id}))

	}

	err = deliveryOrder.GetPreloadDb(false,false, preloads).First(deliveryOrder, deliveryOrder.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (deliveryOrder *DeliveryOrder) delete () error {

	return deliveryOrder.GetPreloadDb(true,false,nil).Where("id = ?", deliveryOrder.Id).Delete(deliveryOrder).Error
}
// ######### END CRUD Functions ############

func (deliveryOrder *DeliveryOrder) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(deliveryOrder)
	} else {
		_db = _db.Model(&DeliveryOrder{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"WebSite","Customer","Status","Order"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}

/* Установка статуса отмена */
func (deliveryOrder *DeliveryOrder) SetCanceledAnyStatus() error {

	if err := deliveryOrder.load(nil); err != nil {
		return err
	}

	cStatus, err := DeliveryStatus{}.GetCanceledAnyStatus()
	if err != nil { return err}

	/* Чтобы лишний раз не вызывать события */
	if cStatus.Id == deliveryOrder.StatusId {
		return  nil
	}
	
	if err := deliveryOrder.update(map[string]interface{}{"status_id": cStatus.Id}, nil); err != nil {
		return err
	}
	
	return nil
}