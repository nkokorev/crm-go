package models

import (
	"database/sql"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"log"
	"time"
)

type Order struct {
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	PublicId	uint   	`json:"public_id" gorm:"type:int;index;not null;"` // Публичный ID заказа внутри магазина
	AccountId 	uint 	`json:"account_id" gorm:"type:int;index;not null"` // аккаунт-владелец ключа
	
	// Ответственный менеджер, назначается внутри системы
	ManagerId 	*uint	`json:"manager_id" gorm:"type:int;"`
	Manager		User	`json:"manager"`

	////// Данные заказа ///////
	Individual	bool 	`json:"individual" gorm:"type:bool;default:true;not null;"` // Физ.лицо - true, Юрлицо - false

	// Комментарий клиента к заказу
	CustomerComment *string	`json:"customer_comment" gorm:"type:varchar(255);"`

	// Магазин (сайт) с которого пришел заказ. НЕ может быть null.
	WebSiteId 	uint	`json:"web_site_id" gorm:"type:int;not null;"`
	WebSite		WebSite	`json:"web_site"`

	// Данные клиента
	CustomerId 	uint	`json:"customer_id" gorm:"type:int;"`
	Customer	User	`json:"customer"`

	// Данные компании-заказчика
	CompanyId 	*uint	`json:"company_id" gorm:"type:int;"`
	Company		User	`json:"company"` //todo: создать компании

	// Способ (канал) заказа: "Заказ из корзины", "Заказ по телефону", "Пропущенный звонок", "Письмо.."
	OrderChannelId 	uint	`json:"order_channel_id" gorm:"type:int;not null;"`
	OrderChannel 	OrderChannel `json:"order_channel" gorm:"preload"`

	//	Выбранный клиентом способ оплаты:
	PaymentMethodId 	uint	`json:"payment_method_id" gorm:"type:int;"`
	PaymentMethodType 	string	`json:"payment_method_type" gorm:"type:varchar(32);default:'payment_yandexes'"`
	PaymentMethod 		PaymentMethod `json:"payment_method" gorm:"-"` // <<< интерфейс

	// Фиксируем стоимость заказа
	AmountId  	uint			`json:"amount_id" gorm:"type:int;"`
	Amount		PaymentAmount	`json:"amount"`

	// Состав заказа
	CartItems	[]CartItem	`json:"cart_items"`
	Payment		Payment		`json:"payment" gorm:"preload"`

	// Данные о доставке
	// DeliveryOrderId	*uint	`json:"deliveryOrderId" gorm:"type:int;"`
	DeliveryOrder	DeliveryOrder	`json:"delivery_order"`

	// Комментарии менеджеров к заказу
	Comments	[]Comment `json:"comments" gorm:"many2many:order_comment;"`

	// Статус заказа
	StatusId  	uint		`json:"status_id" gorm:"type:int;default:1;"`
	Status		OrderStatus	`json:"status" gorm:"preload"`

	// IpV4	Ipv

	CreatedAt time.Time 	`json:"created_at"`
	UpdatedAt time.Time 	`json:"updated_at"`
	DeletedAt *time.Time 	`json:"deleted_at"`
}

func (Order) PgSqlCreate() {
	if err := db.Migrator().AutoMigrate(&Order{});err != nil {log.Fatal(err)}
	// db.Model(&Order{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&Order{}).AddForeignKey("amount_id", "payment_amounts(id)", "RESTRICT", "CASCADE")
	// db.Model(&Order{}).AddForeignKey("order_channel_id", "order_channels(id)", "RESTRICT", "CASCADE")
	// db.Model(&Order{}).AddForeignKey("status_id", "order_statuses(id)", "RESTRICT", "CASCADE")
	err := db.Exec("ALTER TABLE orders " +
		"ADD CONSTRAINT orders_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT orders_amount_id_fkey FOREIGN KEY (amount_id) REFERENCES payment_amounts(id) ON DELETE RESTRICT ON UPDATE CASCADE," +
		"ADD CONSTRAINT orders_order_channel_id_fkey FOREIGN KEY (order_channel_id) REFERENCES order_channels(id) ON DELETE RESTRICT ON UPDATE CASCADE," +
		"DROP CONSTRAINT IF EXISTS fk_orders_customer," +
		"DROP CONSTRAINT IF EXISTS fk_orders_manager," +
		"DROP CONSTRAINT IF EXISTS fk_orders_company," +
		"ADD CONSTRAINT orders_status_id_fkey FOREIGN KEY (status_id) REFERENCES order_statuses(id) ON DELETE RESTRICT ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (order *Order) BeforeCreate(tx *gorm.DB) error {
	order.Id = 0

	// 1. Рассчитываем PublicId (#id заказа) внутри аккаунта
	var lastIdx sql.NullInt64
	err := db.Model(&Order{}).Where("account_id = ?",  order.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	order.PublicId = 1 + uint(lastIdx.Int64)

	// 2. Fix accountId in Amount
	order.Amount.AccountId = order.AccountId
	
	return nil
}
func (order *Order) AfterCreate(tx *gorm.DB) (error) {
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
func (order *Order) AfterFind(tx *gorm.DB) (err error) {

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

	// fmt.Println(wb.ManagerId)
	// fmt.Println(wb.Manager)
	if err := db.Create(&wb).Error; err != nil {
		return nil, err
	}
	if err := wb.GetPreloadDb(false,false, true).First(&wb,wb.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &wb

	return entity, nil
}
func (Order) get(id uint, preloads []string) (Entity, error) {

	var order Order

	err := (&Order{}).GetPreloadDb(false,false, true).
		First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}
func (order *Order) load(preloads []string) error {


	if order.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить Order - не указан  Id"}
	}

	if err := order.GetPreloadDb(false,true, true).First(order, order.Id).Error; err != nil {
		return err
	}

	return nil
}
func (order *Order) loadByPublicId(preloads []string) error {

	if order.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить Order - не указан  Id"}
	}

	if err := order.GetPreloadDb(false,false, true).First(order, "account_id = ? AND public_id = ?", order.AccountId, order.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (Order) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return Order{}.getPaginationList(accountId, 0,100,sortBy,"",nil,preload)
}
func (Order) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {
	
	orders := make([]Order,0)
	var total int64

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
		// err := db.Model(&Order{}).Preload("Status").Preload("Payment").
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
	for i := range orders {
		entities[i] = &orders[i]
	}

	return entities, total, nil
}
func (order *Order) update(input map[string]interface{}, preloads []string) error {

	delete(input,"webSite")
	delete(input,"orderChannel")
	delete(input,"comments")
	delete(input,"manager")
	delete(input,"products")
	delete(input,"company")
	delete(input,"client")
	delete(input,"cartItems")

	err := order.GetPreloadDb(true,false, false).Where("id = ?", order.Id).
		Omit("id", "account_id").Updates(input).Error
	if err != nil { return err }

	_ = order.load(preloads []string)
	return nil
}
func (order *Order) delete () error {

	var idx = make([]uint,0)
	idx = append(idx,order.AmountId)

	if err := (PaymentAmount{}).deletes(idx); err != nil {
		return err
	}

	if err := order.Amount.delete(); err != nil {
		return err
	}

	return db.Where("id = ?", order.Id).Delete(order).Error
}
// ######### END CRUD Functions ############

// ########## Work function ############
func (order *Order) AppendProducts (products []Product) error {
	if err := db.Model(order).Association("CartItems").Replace(products); err != nil {
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
			Preload("Amount").Preload("CartItems").	Preload("CartItems.Product").Preload("CartItems.Amount").Preload("CartItems.PaymentMode").
			Preload("Manager").Preload("WebSite").Preload("OrderChannel")
	} else {
		return _db
	}


}

