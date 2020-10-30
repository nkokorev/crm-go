package models

import (
	"database/sql"
	"errors"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

	Individual	bool 	`json:"individual" gorm:"type:bool;default:true;not null;"` // Физ.лицо - true, Юрлицо - false

	// Комментарий клиента к заказу
	CustomerComment *string	`json:"customer_comment" gorm:"type:varchar(255);"`

	// Магазин (сайт) с которого пришел заказ. НЕ может быть null.
	WebSiteId 	*uint	`json:"web_site_id" gorm:"type:int;not null;"`
	WebSite		WebSite	`json:"web_site"`

	// Данные клиента
	CustomerId 	*uint	`json:"customer_id" gorm:"type:int;"` 	// << null при новом клиенте или не опознанном
	Customer	User	`json:"customer"`

	// Данные компании-заказчика
	CompanyId 	*uint	`json:"company_id" gorm:"type:int;"` 	// << null при новой компании или не опознанной
	Company		Company	`json:"company"` //todo: создать компании

	// Способ (канал) заказа: "Заказ из корзины", "Заказ по телефону", "Пропущенный звонок", "Письмо.."
	OrderChannelId 	uint	`json:"order_channel_id" gorm:"type:int;not null;"`
	OrderChannel 	OrderChannel `json:"order_channel" gorm:"preload"`

	//	Выбранный клиентом способ оплаты:
	PaymentMethodId 	uint	`json:"payment_method_id" gorm:"type:int;"`
	PaymentMethodType 	string	`json:"payment_method_type" gorm:"type:varchar(32);default:'payment_yandexes'"`
	PaymentMethod 		PaymentMethod `json:"payment_method" gorm:"-"` // <<< интерфейс

	// Фиксируем стоимость заказа
	Amount		PaymentAmount	`json:"amount" gorm:"-"`
	Cost		float64 	`json:"cost" gorm:"type:numeric;default:0"`

	// Состав заказа
	CartItems	[]CartItem	`json:"cart_items"`
	Payment		Payment		`json:"payment"`

	// Данные о доставке
	// DeliveryOrderId	*uint	`json:"deliveryOrderId" gorm:"type:int;"`
	DeliveryOrder	*DeliveryOrder	`json:"delivery_order"`

	// Комментарии менеджеров к заказу
	Comments	[]Comment `json:"comments" gorm:"many2many:order_comment;"`

	// Статус заказа
	StatusId  	uint		`json:"status_id" gorm:"type:int;default:1;"`
	Status		OrderStatus	`json:"status"`

	// IpV4	Ipv

	CreatedAt time.Time 	`json:"created_at"`
	UpdatedAt time.Time 	`json:"updated_at"`
	DeletedAt *time.Time 	`json:"deleted_at"`
}

func (Order) PgSqlCreate() {
	if err := db.Migrator().AutoMigrate(&Order{});err != nil {log.Fatal(err)}
	// db.Model(&Order{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&Order{}).AddForeignKey("order_channel_id", "order_channels(id)", "RESTRICT", "CASCADE")
	// db.Model(&Order{}).AddForeignKey("status_id", "order_statuses(id)", "RESTRICT", "CASCADE")
	err := db.Exec("ALTER TABLE orders " +
		"ADD CONSTRAINT orders_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT orders_order_channel_id_fkey FOREIGN KEY (order_channel_id) REFERENCES order_channels(id) ON DELETE RESTRICT ON UPDATE CASCADE," +
		"DROP CONSTRAINT IF EXISTS fk_orders_customer," +
		"DROP CONSTRAINT IF EXISTS fk_orders_manager," +
		"DROP CONSTRAINT IF EXISTS fk_orders_company," +
		"ADD CONSTRAINT orders_status_id_fkey FOREIGN KEY (status_id) REFERENCES order_statuses(id) ON DELETE RESTRICT ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (order *Order) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(order)
	} else {
		_db = _db.Model(&Order{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,
			[]string{"Status","Payment","Customer","DeliveryOrder","DeliveryOrder.Status","DeliveryOrder.Status","DeliveryOrder.Amount",
			"CartItems","CartItems.Product","CartItems.Warehouse","CartItems.Product.MeasurementUnit","CartItems.Product.ProductCards","CartItems.Product.ProductTags","CartItems.Product.PaymentSubject",
			"CartItems.Amount","CartItems.PaymentMode","Manager","WebSite","OrderChannel","Company","Comments"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
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

	return nil
}
func (order *Order) AfterCreate(tx *gorm.DB) error {
	// AsyncFire(*Event{}.OrderCreated(order.AccountId, order.Id))
	AsyncFire(NewEvent("OrderCreated", map[string]interface{}{"account_id":order.AccountId, "order_id":order.Id}))
	return nil
}
func (order *Order) AfterUpdate(tx *gorm.DB) (err error) {
	// AsyncFire(*Event{}.OrderUpdated(order.AccountId, order.Id))
	AsyncFire(NewEvent("OrderUpdated", map[string]interface{}{"account_id":order.AccountId, "order_id":order.Id}))

	orderStatusEntity, err := OrderStatus{}.get(order.StatusId,nil)
	if err == nil && orderStatusEntity.GetAccountId() == order.AccountId {
		if orderStatus, ok := orderStatusEntity.(*OrderStatus); ok {
			if orderStatus.Code == "completed" {
				// AsyncFire(*Event{}.OrderCompleted(order.AccountId, order.Id))
				AsyncFire(NewEvent("OrderCompleted", map[string]interface{}{"account_id":order.AccountId, "order_id":order.Id}))
			}
			if orderStatus.Code == "canceled" {
				// AsyncFire(*Event{}.OrderCanceled(order.AccountId, order.Id))
				AsyncFire(NewEvent("OrderCanceled", map[string]interface{}{"account_id":order.AccountId, "order_id":order.Id}))
			}
		}

	}
	return nil
}
func (order *Order) AfterDelete(tx *gorm.DB) (err error) {
	// AsyncFire(*Event{}.OrderDeleted(order.AccountId, order.Id))
	AsyncFire(NewEvent("OrderDeleted", map[string]interface{}{"account_id":order.AccountId, "order_id":order.Id}))
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
		err := Account{Id: order.AccountId}.LoadEntity(&paymentSubject, order.CartItems[i].PaymentSubjectId,nil)
		if err != nil { return err}
		order.CartItems[i].PaymentSubjectYandex = paymentSubject.Code
	}

	order.Amount.Value = order.Cost
	order.Amount.Currency = "RUB"

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
	
	_item := order
	// if err := db.Omit(clause.Associations).Select("Amount").Create(&_item).Error; err != nil {
	if err := db.Omit("Customer","Manager","WebSite","Company","OrderChannel","PaymentMethod","Payment","DeliveryOrder","Comments","Status").
		Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}
func (Order) get(id uint, preloads []string) (Entity, error) {

	var order Order

	err := (&Order{}).GetPreloadDb(false,false, preloads).
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

	if err := order.GetPreloadDb(false,false, preloads).First(order, order.Id).Error; err != nil {
		return err
	}

	return nil
}
func (order *Order) loadByPublicId(preloads []string) error {

	if order.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить Order - не указан  Id"}
	}

	if err := order.GetPreloadDb(false,false, preloads).First(order, "account_id = ? AND public_id = ?", order.AccountId, order.PublicId).Error; err != nil {
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

		err := (&Order{}).GetPreloadDb(false, false, preloads).
			Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&orders, "customer_comment ILIKE ?", search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&Order{}).GetPreloadDb(false, false, nil).
			Where("account_id = ? AND customer_comment ILIKE ?", accountId, search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&Order{}).GetPreloadDb(false, false, preloads).
		// err := db.Model(&Order{}).Preload("Status").Preload("Payment").
				Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&orders).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&Order{}).GetPreloadDb(false, false, nil).Where("account_id = ?", accountId).Count(&total).Error
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

	delete(input,"status")
	delete(input,"payment")
	delete(input,"customer")
	delete(input,"delivery_order")
	delete(input,"cart_items")
	delete(input,"manager")
	delete(input,"web_site")
	delete(input,"order_channel")
	delete(input,"company")
	delete(input,"comments")
	delete(input,"payment_method")
	delete(input,"amount")

	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id","manager_id","web_site_id","customer_id","company_id","order_channel_id","payment_method_id",
		"payment_method_id","status_id"}); err != nil {
		return err
	}

	// Отметка на будущее для события
	_newStatusId, ok := input["status_id"].(uint)
	_oldStatusId :=  order.StatusId

	if err := order.GetPreloadDb(false, false, nil).Where("id = ?", order.Id).
		Omit("id", "account_id","public_id").Updates(input).Error; err != nil {return err}

	err := order.GetPreloadDb(false,true, preloads).First(order, order.Id).Error
	if err != nil {
		return err
	}

	// Если флаг статуса заказа был изменен
	if ok && (_newStatusId != uint(_oldStatusId)) {

		switch order.Status.Group {
		case "agreement":
			if order.Status.Code == "agreement_order" {
				AsyncFire(NewEvent("OrderConfirmed", map[string]interface{}{"account_id":order.AccountId, "order_id":order.Id}))
			}
			if order.Status.Code == "agreement_change" {
				AsyncFire(NewEvent("OrderChanging", map[string]interface{}{"account_id":order.AccountId, "order_id":order.Id}))
			}
			if order.Status.Code == "agreement_approval" {
				AsyncFire(NewEvent("OrderApproving", map[string]interface{}{"account_id":order.AccountId, "order_id":order.Id}))
			}

		case "equipping":
			if order.Status.Code == "equipping" {
				AsyncFire(NewEvent("OrderEquipping", map[string]interface{}{"account_id":order.AccountId, "order_id":order.Id}))
			}
			if order.Status.Code == "equipped" {
				AsyncFire(NewEvent("OrderEquipped", map[string]interface{}{"account_id":order.AccountId, "order_id":order.Id}))
			}

		case "delivery":
			if order.Status.Code == "delivery_sent" {
				AsyncFire(NewEvent("OrderDeliverySent", map[string]interface{}{"account_id":order.AccountId, "order_id":order.Id}))
			}
			if order.Status.Code == "delivering" {
				AsyncFire(NewEvent("OrderInDeliveryProcess", map[string]interface{}{"account_id":order.AccountId, "order_id":order.Id}))
			}
			if order.Status.Code == "delivery_rescheduled" {
				AsyncFire(NewEvent("OrderDeliveryRescheduled", map[string]interface{}{"account_id":order.AccountId, "order_id":order.Id}))
			}
			
		case "completed":
			if err := order.SetWastedOffFromWarehouse(); err != nil { return err}
			AsyncFire(NewEvent("OrderCompleted", map[string]interface{}{"account_id":order.AccountId, "order_id":order.Id}))

		case "prepend":
			AsyncFire(NewEvent("OrderPrepending", map[string]interface{}{"account_id":order.AccountId, "order_id":order.Id}))

		case "canceled":
			var deliveryOrder DeliveryOrder
			err := Account{Id: order.AccountId}.LoadEntity(&deliveryOrder, order.DeliveryOrder.Id, []string{"DeliveryOrder"})
			if err != nil {
				log.Println("Error get delivery Order in canceled order: ", err)
			} else {
				if err := deliveryOrder.SetCanceledAnyStatus(); err != nil {log.Println("SetCanceledAnyStatus deliveryOrder: ", err)}
			}

			AsyncFire(NewEvent("OrderCanceled", map[string]interface{}{"account_id":order.AccountId, "order_id":order.Id}))
		}

		// общая информация об обновлении статуса
		AsyncFire(NewEvent("OrderStatusUpdated", map[string]interface{}{"account_id":order.AccountId, "order_id":order.Id}))

	}
	
	err = order.GetPreloadDb(false,false, preloads).First(order, order.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (order *Order) delete () error {

	// 1. Удаляем Amount
	if err := (&PaymentAmount{Id: order.AccountId}).delete(); err != nil {
		return err
	}

	return order.GetPreloadDb(true,false,nil).Where("id = ?", order.Id).Delete(order).Error
}
// ######### END CRUD Functions ############

// ########## Work function ############
func (order *Order) AppendProducts (products []Product) error {
	if err := db.Model(order).Association("CartItems").Replace(products); err != nil {
		return err
	}

	return nil
}

/* Изменение статуса заказа на "Выполнено" */
func (order *Order) SetCompleted () error {

	if order.Id < 0 {
		if err := order.load(nil);err != nil {
			return err
		}
	}


	// Получаем статус "Выполнено", для извлечения его ID
	cStatus, err := (OrderStatus{}).GetCompletedStatus()
	if err != nil { return err}

	/* Чтобы лишний раз не вызывать событие "Выполнено" */
	if cStatus.Id != order.StatusId {
		if err := order.update(map[string]interface{}{"status_id": cStatus.Id},nil); err != nil {
			return err
		}

		if err := order.SetWastedOffFromWarehouse(); err != nil { return err}
	}

	// Перевод товаров из резерва в order
	
	return nil
}

// Переводит все товарные позиции в статус "Выполнено"
func (order Order) SetWastedOffFromWarehouse() error {

	// 1. Загружаем все CartItems

	cartItems := make([]CartItem,0)
	err := db.Model(&CartItem{}).Where("account_id = ? AND order_id = ? AND product_id > 0 AND reserved = true", order.AccountId, order.Id).
		Find(&cartItems).Error
	if err != nil {	return err	}

	for i := range cartItems {
		if err := cartItems[i].SetWastedOffFromWarehouse(); err != nil {
			log.Printf("Ошибка списания cartItem id =[%v] со склада: %v \n", cartItems[i].Id, err)
		}
	}

	return nil
}

/* Изменение статуса заказа на "Отмена" */
func (order *Order) SetCanceled () error {

	if err := order.load(nil);err != nil {
		return err
	}

	// Получаем статус "Выполнено", для извлечения его ID
	cStatus, err := (OrderStatus{}).GetCanceledAnyStatus()
	if err != nil { return err}

	/* Чтобы лишний раз не вызывать событие "Выполнено" */
	if cStatus.Id != order.StatusId {
		if err := order.update(map[string]interface{}{"status_id": cStatus.Id},nil); err != nil {
			return err
		}
	}

	return nil
}

// Добавить/Удалить позицию товара
func (order *Order) AppendProduct(product Product, strict... bool) error {

	// 1. Загружаем продукт еще раз
	if err := product.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: нельзя добавить продукт, т.к. он не найден"}
	}

	// 2. Проверяем есть ли уже на этом складе этот продукт
	if order.ExistProduct(product.Id) {
		if len(strict) > 1 {
			return utils.Error{Message: "Продукт уже числиться на складе"}
		} else {
			return nil
		}
	}

	shortDesc := ""
	if product.ShortDescription != nil {
		shortDesc = *product.ShortDescription
	} else {
		if product.Label != nil {
			shortDesc = *product.Label
		}
	}
	cost := float64(0)
	if product.RetailPrice != nil {

		if product.RetailDiscount != nil {
			cost = *product.RetailPrice - *product.RetailDiscount
		} else {
			cost = *product.RetailPrice
		}
	}

	paymentSubjectId := uint(0)
	if product.PaymentSubjectId != nil {
		paymentSubjectId = *product.PaymentSubjectId
	}

	VatCodeId := uint(0)
	if product.VatCodeId != nil {
		VatCodeId = *product.VatCodeId
	}

	if err := db.Model(&CartItem{}).Create(
		&CartItem{
			AccountId: order.AccountId,
			ProductId: product.Id,
			OrderId: order.Id,
			Description: shortDesc,
			Cost: cost,
			PaymentSubjectId: paymentSubjectId,
			VatCode:VatCodeId,
		}).Error; err != nil {
		return err
	}

	account, err := GetAccount(order.AccountId)
	if err == nil && account != nil {
		// AsyncFire(*Event{}.OrderProductAppended(account.Id, order.Id, product.Id))
	}

	_ = order.UpdateDeliveryData()

	return nil
}
func (order *Order) RemoveProduct(product Product) error {

	// 1. Загружаем продукт еще раз
	if err := product.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: нельзя удалить продукт, он не найден"}
	}

	if order.AccountId < 1 || product.Id < 1 || order.Id < 1 {
		return utils.Error{Message: "Техническая ошибка: account id || product id || order id == nil"}
	}

	var cartItem CartItem
	if err := db.Model(&CartItem{}).Where("account_id = ? AND product_id = ? AND order_id = ?", order.AccountId, product.Id, order.Id).First(&cartItem).Error; err != nil {
		return err
	}
	if err := cartItem.delete(); err != nil {
		return err
	}

	/*if err := db.Where("account_id = ? AND product_id = ? AND order_id = ?", order.AccountId, product.Id, order.Id).Delete(
		&CartItem{}).Error; err != nil {
		return err
	}*/

	_ = order.UpdateDeliveryData()

	return nil
}
func (order *Order) ExistProduct(productId uint) bool {

	if productId < 1 {
		return false
	}

	var count int64

	db.Model(&CartItem{}).Where("order_id = ? AND product_id = ?", order.Id, productId).Count(&count)
	if count > 0 {
		return true
	}

	return false
}

func (order *Order) AppendDeliveryItem(delivery Delivery, cost float64) error {

	// 1. Загружаем продукт еще раз
	if err := delivery.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: нельзя добавить продукт, т.к. он не найден"}
	}

	vatCode, err :=  delivery.GetVatCode()
	if err != nil { return err }

	if err := db.Model(&CartItem{}).Create(
		&CartItem{
			AccountId: order.AccountId,
			ProductId: 0, // т.к. это услуга
			OrderId: order.Id,
			Description: delivery.GetName(),
			Cost: cost,
			PaymentSubjectId: delivery.GetPaymentSubject().Id,
			PaymentSubjectYandex: delivery.GetPaymentSubject().Code,
			VatCode: vatCode.Id,
			Quantity: 1,
		}).Error; err != nil {
		return err
	}

	_ = order.UpdateDeliveryData()

	return nil
}
func (order *Order) UpdateDeliveryItem(input map[string]interface{}) error {

	if order.AccountId < 1 || order.Id < 1 {
		return utils.Error{Message: "Техническая ошибка: account id || order id == nil"}
	}

	// Ищем CartItem
	var cartItem CartItem
	if err := db.Where("account_id = ? AND product_id = 0 AND order_id = ?", order.AccountId, order.Id).First(&cartItem).Error; err != nil {
		return err
	}

	if err := cartItem.update(input, nil); err != nil {
		return err
	}

	_ = order.UpdateDeliveryData()

	return nil
}
func (order *Order) RemoveDeliveryItem() error {

	if order.AccountId < 1 || order.Id < 1 {
		return utils.Error{Message: "Техническая ошибка: account id || order id == nil"}
	}

	if err := db.Where("account_id = ? AND product_id = 0 AND order_id = ?", order.AccountId, order.Id).Delete(
		&CartItem{}).Error; err != nil {
		return err
	}


	return nil
}
func (order *Order) ExistDeliveryItem() bool {

	if order.Id < 1 {
		return false
	}

	var count int64

	db.Model(&CartItem{}).Where("order_id = ? AND product_id = 0 AND account_id = ?", order.Id, order.AccountId).Count(&count)
	if count > 0 {
		return true
	}

	return false

}

// Обновляет связанные данные в доставке
func (order Order) UpdateDeliveryData() error {

	if order.AccountId < 1 || order.Id < 1 {
		return utils.Error{Message: "Техническая ошибка: account id || order id == nil"}
	}

	// Загружаем еще раз заказ
	if err := order.load([]string{"CartItems.Product"}); err != nil { return err}
	
	// Ищем CartItem
	var deliveryOrder DeliveryOrder
	if err := db.Where("account_id = ? AND order_id = ?", order.AccountId, order.Id).First(&deliveryOrder).Error; err != nil {
		return nil
	}
	if deliveryOrder.Id < 1 {
		return nil
	}

	weight := float64(0)
	for _, v := range order.CartItems {
		if v.ProductId > 0 && v.Product.Weight != nil {
			weight += v.Quantity * *v.Product.Weight
		}
	}

	if err := deliveryOrder.update(map[string]interface{}{"weight":weight}, nil); err != nil {
		return err
	}

	return nil

}
func (order Order) UpdateCost() error {

	if err := order.load([]string{"CartItems"}); err != nil {
		return err
	}
	cost := float64(0)

	for i := range order.CartItems {
		cost += order.CartItems[i].Quantity * order.CartItems[i].Cost
	}
	if err := order.update(map[string]interface{}{"cost":cost}, nil); err != nil { return err }

	return nil
}

// Устанавливает неизвестного заказчика
func (order *Order) SetUnknownCustomer() error {
	if order.Id < 1 || order.AccountId < 1 {
		return errors.New("Техническая ошибка установки неизвестного заказчика")
	}

	if err := order.load([]string{"Customer"}); err != nil { return err}
	if order.Customer.IsUnknown {
		return nil
	}

	account, err := GetAccount(order.AccountId)
	if err != nil { return err }

	// Находим ID неизвестного пользователя
	user, err := account.FindOrCreateUnknownUser()
	if err != nil { return nil }

	// Обновляем customer_id
	if err := order.update(map[string]interface{}{"customer_id":user.Id}, nil); err != nil { return nil }

	return nil
}

