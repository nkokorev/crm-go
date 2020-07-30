package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type DeliveryOrder struct {
	Id     			uint   	`json:"id" gorm:"primary_key"`
	PublicId	uint   	`json:"publicId" gorm:"type:int;index;not null;"` // Публичный ID заказа внутри магазина
	AccountId 		uint 	`json:"-" gorm:"type:int;index;not null;"`

	// Данные заказа
	OrderId 	uint	`json:"orderId" gorm:"type:int;not null"`
	Order	Order	`json:"order"`

	// Данные заказчика
	CustomerId 	uint	`json:"customerId" gorm:"type:int;not null"`
	Customer	User	`json:"customer"`

	// Магазин (сайт) с которого пришел заказ. НЕ может быть null.
	WebSiteId 	uint	`json:"webSiteId" gorm:"type:int;not null;"`
	WebSite		WebSite	`json:"webSite"`

	// Тело заказа
	Code	string 	`json:"deliveryCode" gorm:"type:varchar(32);"`
	MethodId 		uint	`json:"methodId" gorm:"type:int;not null;"`

	Address	string 	`json:"deliveryAddress" gorm:"type:varchar(32);"`
	PostalCode	string 	`json:"deliveryPostalCode" gorm:"type:varchar(32);"`
	Delivery	Delivery	`json:"delivery" gorm:"-"` // << preload

	// Фиксируем стоимость
	AmountId  	uint			`json:"amountId" gorm:"type:int;not null;"`
	Amount  	PaymentAmount	`json:"amount"`

	// Статус заказа
	DeliveryStatusId  	uint	`json:"deliveryStatusId" gorm:"type:int;not null;"`
	DeliveryStatus		DeliveryStatus	`json:"deliveryStatus"`

	CreatedAt 		time.Time `json:"createdAt"`
	UpdatedAt 		time.Time `json:"updatedAt"`
}

// ############# Entity interface #############
func (deliveryOrder DeliveryOrder) GetId() uint { return deliveryOrder.Id }
func (deliveryOrder *DeliveryOrder) setId(id uint) { deliveryOrder.Id = id }
func (deliveryOrder DeliveryOrder) GetAccountId() uint { return deliveryOrder.AccountId }
func (deliveryOrder *DeliveryOrder) setAccountId(id uint) { deliveryOrder.AccountId = id }
func (DeliveryOrder) SystemEntity() bool { return false }

// ############# Entity interface #############

func (DeliveryOrder) PgSqlCreate() {
	db.AutoMigrate(&DeliveryOrder{})
	db.Model(&DeliveryOrder{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&DeliveryOrder{}).AddForeignKey("order_id", "orders(id)", "CASCADE", "CASCADE")
	db.Model(&DeliveryOrder{}).AddForeignKey("delivery_status_id", "delivery_statuses(id)", "CASCADE", "CASCADE")
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
	event.AsyncFire(Event{}.DeliveryOrderCreated(deliveryOrder.AccountId, deliveryOrder.Id))
	return nil
}
func (deliveryOrder *DeliveryOrder) AfterUpdate(tx *gorm.DB) (err error) {

	event.AsyncFire(Event{}.DeliveryOrderUpdated(deliveryOrder.AccountId, deliveryOrder.Id))

	orderStatusEntity, err := DeliveryStatus{}.get(deliveryOrder.DeliveryStatusId)
	if err == nil && orderStatusEntity.GetAccountId() == deliveryOrder.AccountId {
		if deliveryStatus, ok := orderStatusEntity.(*DeliveryStatus); ok {
			if deliveryStatus.Code == "completed" {
				event.AsyncFire(Event{}.DeliveryOrderCompleted(deliveryOrder.AccountId, deliveryOrder.Id))
			}
			if deliveryStatus.Code == "canceled" {
				event.AsyncFire(Event{}.DeliveryOrderCanceled(deliveryOrder.AccountId, deliveryOrder.Id))
			}
		}

	}

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

	err := db.Preload("WebSite").Preload("Amount").Preload("Order").Preload("Customer").First(deliveryOrder, deliveryOrder.Id).Error
	if err != nil {
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

		err := db.Model(&DeliveryOrder{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Preload("WebSite").Preload("Amount").Preload("Order").Preload("Customer").
			Find(&deliveryOrders, "name ILIKE ? OR description ILIKE ?", search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&DeliveryOrder{}).
			Where("account_id = ? AND name ILIKE ? OR description ILIKE ? ", accountId, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&DeliveryOrder{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Preload("WebSite").Preload("Amount").Preload("Order").Preload("Customer").
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

	// Приводим в опрядок
	input = utils.FixJSONB_String(input, []string{"recipientList"})
	input = utils.FixJSONB_Uint(input, []string{"recipientUsersList"})

	delete(input, "order")

	// work!!!
	if err := db.Set("gorm:association_autoupdate", false).Model(DeliveryOrder{}).Where(" id = ?", deliveryOrder.Id).
		Omit("id", "account_id","created_at").Updates(input).Error; err != nil {
		return err
	}

	err := db.Preload("WebSite").Preload("Amount").Preload("Order").Preload("Customer").First(deliveryOrder, deliveryOrder.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (deliveryOrder DeliveryOrder) delete () error {
	return db.Model(DeliveryOrder{}).Where("id = ?", deliveryOrder.Id).Delete(deliveryOrder).Error
}
// ######### END CRUD Functions ############
