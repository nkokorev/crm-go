package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type Order struct {
	Id     		uint   	`json:"id" gorm:"primary_key"`
	PublicId	uint   	`json:"publicId" gorm:"type:int;index;not null;"` // Публичный ID заказа внутри магазина
	AccountId 	uint 	`json:"accountId" gorm:"type:int;index;not null"` // аккаунт-владелец ключа
	
	// Ответственный менеджер, назначается внутри системы
	ManagerId 	uint	`json:"managerId" gorm:"type:int;not null"`
	Manager		User	`json:"manager"`

	////// Данные заказа ///////
	Individual	bool 	`json:"individual" gorm:"type:bool;default:true;not null;"` // Физ.лицо - true, Юрлицо - false

	// Комментарий клиента к заказу
	CustomerComment string	`json:"customerComment" gorm:"type:varchar(255);"`

	// Магазин (сайт) с которого пришел заказ. НЕ может быть null.
	WebSiteId 	uint	`json:"webSiteId" gorm:"type:int;not null;"`
	WebSite		WebSite	`json:"webSite"`

	// Данные клиента
	CustomerId 	uint	`json:"customerId" gorm:"type:int;not null"`
	Customer	User	`json:"customer"`

	// Данные компании-заказчика
	CompanyId 	uint	`json:"companyId" gorm:"type:int;not null"`
	Company		User	`json:"company"`

	// Способ (канал) заказа: "Заказ из корзины", "Заказ по телефону", "Пропущенный звонок", "Письмо.."
	OrderChannelId 	uint	`json:"orderChannelId" gorm:"type:int;not null;"`
	OrderChannel 	OrderChannel `json:"orderChannel"`

	//	Выбранный клиентом способ оплаты:
	PaymentMethodId 	uint	`json:"paymentMethodId" gorm:"type:int;not null;default:1"`
	PaymentMethodType 	string	`json:"paymentMethodType" gorm:"type:varchar(32);not null;default:'payment_yandexes'"`
	PaymentMethod 		PaymentMethod `json:"paymentMethod" gorm:"-"`

	// Фиксируем стоимость заказа
	AmountId  	uint	`json:"amountId" gorm:"type:int;not null;"`
	Amount		PaymentAmount	`json:"amount"`

	// Состав заказа
	CartItems	[]CartItem		`json:"cartItems"`
	Payment	Payment	`json:"payment"`

	// Данные о доставке
	// DeliveryOrderId	*uint	`json:"deliveryOrderId" gorm:"type:int;"`
	DeliveryOrder	DeliveryOrder	`json:"deliveryOrder"`

	// Комментарии менеджеров к заказу
	Comments	[]OrderComment `json:"comments"`

	// Статус заказа
	StatusId  	uint	`json:"statusId" gorm:"type:int;default:1;"`
	Status		OrderStatus	`json:"status"`

	CreatedAt time.Time 	`json:"createdAt"`
	UpdatedAt time.Time 	`json:"updatedAt"`
	DeletedAt *time.Time 	`json:"deletedAt"`
}

func (Order) PgSqlCreate() {
	db.AutoMigrate(&Order{})
	db.Model(&Order{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&Order{}).AddForeignKey("amount_id", "payment_amounts(id)", "CASCADE", "CASCADE")
	db.Model(&Order{}).AddForeignKey("order_channel_id", "order_channels(id)", "CASCADE", "CASCADE")
	db.Model(&Order{}).AddForeignKey("status_id", "order_statuses(id)", "CASCADE", "CASCADE")

	// fmt.Println("Щквук!")
}
func (order *Order) BeforeCreate(scope *gorm.Scope) error {
	order.Id = 0

	// 1. Рассчитываем PublicId (#id заказа) внутри аккаунта
	lastIdx := uint(0)
	var ord Order

	err := db.Where("account_id = ?", order.AccountId).Select("public_id").Last(&ord).Error;
	if err != nil && err != gorm.ErrRecordNotFound { return err}
	if err == gorm.ErrRecordNotFound {
		lastIdx = 0
	} else {
		lastIdx = ord.PublicId
	}
	order.PublicId = lastIdx + 1

	// 2. Fix accountId in Amount
	order.Amount.AccountId = order.AccountId
	
	return nil
}
func (order *Order) AfterCreate(scope *gorm.Scope) (error) {
	event.AsyncFire(Event{}.OrderCreated(order.AccountId, order.Id))
	return nil
}
func (order *Order) AfterUpdate(tx *gorm.DB) (err error) {
	event.AsyncFire(Event{}.OrderUpdated(order.AccountId, order.Id))

	orderStatusEntity, err := OrderStatus{}.get(order.StatusId)
	if err == nil && orderStatusEntity.GetAccountId() == order.AccountId {
		if orderStatus, ok := orderStatusEntity.(*OrderStatus); ok {
			if orderStatus.Code == "completed" {
				event.AsyncFire(Event{}.OrderCompleted(order.AccountId, order.Id))
			}
			if orderStatus.Code == "canceled" {
				event.AsyncFire(Event{}.OrderCanceled(order.AccountId, order.Id))
			}
		}

	}
	return nil
}
func (order *Order) AfterDelete(tx *gorm.DB) (err error) {
	event.AsyncFire(Event{}.OrderDeleted(order.AccountId, order.Id))
	return nil
}
func (order *Order) AfterFind() (err error) {

	if order.PaymentMethodType != "" && order.PaymentMethodId > 0 {
		// Get ALL Payment Methods
		method, err := Account{Id: order.AccountId}.GetPaymentMethodByType(order.PaymentMethodType, order.PaymentMethodId)
		if err != nil { return err }
		order.PaymentMethod = method
	}

	for i := range order.CartItems {
		var paymentSubject PaymentSubject
		err := Account{Id: order.AccountId}.LoadEntity(&paymentSubject, order.CartItems[i].PaymentSubjectId)
		if err != nil { return err}
		order.CartItems[i].PaymentSubjectYandex = paymentSubject.Code
	}

	return nil
}

// ############# Entity interface #############
func (order Order) GetId() uint { return order.Id }
func (order *Order) setId(id uint) { order.Id = id }
func (order *Order) setPublicId(id uint) { order.PublicId = id }
func (order Order) GetAccountId() uint { return order.AccountId }
func (order *Order) setAccountId(id uint) { order.AccountId = id }
func (Order) SystemEntity() bool { return false }
// ############# END of Entity interface #############

// ######### CRUD Functions ############
func (order Order) create() (Entity, error)  {

	wb := order

	if err := db.Create(&wb).First(&wb,wb.Id).Error; err != nil {
		return nil, err
	}
	if err := wb.GetPreloadDb(false,false, true).First(&wb,wb.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &wb

	return entity, nil
}
func (Order) get(id uint) (Entity, error) {

	var order Order

	err := (&Order{}).GetPreloadDb(false,false, true).
		First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}
func (order *Order) load() error {


	if order.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить Order - не указан  Id"}
	}

	if err := order.GetPreloadDb(false,true, true).First(order, order.Id).Error; err != nil {
		return err
	}

	return nil
}
func (order *Order) loadByPublicId() error {

	if order.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить Order - не указан  Id"}
	}

	if err := order.GetPreloadDb(false,false, true).First(order, "public_id = ?", order.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (Order) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return Order{}.getPaginationList(accountId, 0,100,sortBy,"")
}
func (Order) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {
	
	orders := make([]Order,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&Order{}).GetPreloadDb(false,false, true).
			Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&orders, "customer_comment ILIKE ?", search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Order{}).
			Where("account_id = ? AND customer_comment ILIKE ?", accountId, search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&Order{}).GetPreloadDb(false,false, true).
				Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
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
	delete(input,"comments")
	delete(input,"manager")
	delete(input,"products")
	delete(input,"company")
	delete(input,"client")
	delete(input,"cartItems")

	return order.GetPreloadDb(true,false, false).Where("id = ?", order.Id).
		Omit("id", "account_id").Updates(input).Error

}
func (order *Order) delete () error {

	if err := order.Amount.delete(); err != nil {
		return err
	}

	return db.Where("id = ?", order.Id).Delete(order).Error
}
// ######### END CRUD Functions ############

// ########## Work function ############
func (order *Order) AppendProducts (products []Product) error {
	if err := db.Model(order).Association("CartItems").Replace(products).Error; err != nil {
		return err
	}

	return nil
}

func (order *Order) GetPreloadDb(autoUpdateOff bool, getModel bool, preload bool) *gorm.DB {
	_db := db

	if autoUpdateOff {
		_db = _db.Set("gorm:association_autoupdate", false)
	}
	if getModel {
		_db = _db.Model(&order)
	} else {
		_db = _db.Model(&Order{})
	}

	if preload {
		return _db.Preload("Status").Preload("Payment").Preload("Customer").Preload("DeliveryOrder").Preload("DeliveryOrder.Amount").
			Preload("Amount").Preload("CartItems").Preload("CartItems.Product").Preload("CartItems.Amount").Preload("CartItems.PaymentMode").
			Preload("Manager").Preload("WebSite").Preload("OrderChannel")
	} else {
		return _db
	}


}

