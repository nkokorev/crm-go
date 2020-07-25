package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type Order struct {
	Id     		uint   	`json:"id" gorm:"primary_key"`
	PublicId	uint   	`json:"publicId" gorm:"type:int;index;not null;"` // Публичный ID заказа внутри магазина
	AccountId 	uint 	`json:"accountId" gorm:"type:int;index;not null"` // аккаунт-владелец ключа

	// Комментарий клиента к заказу
	CustomerComment string	`json:"customerComment" gorm:"type:varchar(255);"`

	// Комментарии к заказу
	Comments	[]OrderComment `json:"comments"`

	// Ответственный менеджер
	ManagerId 	uint	`json:"managerId" gorm:"type:int;not null"`
	Manager		User	`json:"manager"`

	////// Данные заказа ///////

	Individual	bool 	`json:"individual" gorm:"type:bool;default:true;not null;"` // Физ.лицо - true, Юрлицо - false

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

	// Состав заказа
	CartItems	[]CartItem		`json:"cartItems"`

	// Фиксируем стоимость заказа
	AmountId  	uint	`json:"amountId" gorm:"type:int;not null;"`
	Amount		PaymentAmount	`json:"amount"`

	CreatedAt time.Time 	`json:"createdAt"`
	UpdatedAt time.Time 	`json:"updatedAt"`
	DeletedAt *time.Time 	`json:"deletedAt"`
}


func (Order) PgSqlCreate() {
	if !db.HasTable(&Order{}) {
		db.CreateTable(&Order{})
	}
	db.Model(&Order{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&Order{}).AddForeignKey("amount_id", "payment_amounts(id)", "CASCADE", "CASCADE")

	// fmt.Println("Щквук!")
}
func (order *Order) BeforeCreate(scope *gorm.Scope) error {
	order.Id = 0

	// 1. Рассчитываем PublicId внутри магазина
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

	// 2. 
	order.Amount.AccountId = order.AccountId
	
	return nil
}
func (order *Order)  AfterFind() (err error) {

	// 1. Делаем подсчет стоимости заказа
	if err := order.RetailPriceCalculation(); err != nil {
		return err
	}

	return nil
}

// ############# Entity interface #############
func (order Order) GetId() uint { return order.Id }
func (order *Order) setId(id uint) { order.Id = id }
func (order Order) GetAccountId() uint { return order.AccountId }
func (order *Order) setAccountId(id uint) { order.AccountId = id }
func (Order) SystemEntity() bool { return false }

// ############# Entity interface #############

// ######### CRUD Functions ############
func (order Order) create() (Entity, error)  {

	/**
	Заказ создает в два этапа:

	1. Создается голый заказ c Amount
	3. Добавляются в него товары

	*/

	// fix Amount

	wb := order
	if err := db.Create(&wb).Error; err != nil {
		return nil, err
	}

	var entity Entity = &wb

	return entity, nil
}
func (Order) get(id uint) (Entity, error) {

	var order Order

	err := db.Preload("Amount").Preload("CartItems").Preload("CartItems.Product").Preload("CartItems.Product.PaymentSubject").Preload("CartItems.Amount").Preload("Manager").Preload("WebSite").Preload("OrderChannel").First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}
func (order *Order) load() error {


	if order.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить Order - не указан  Id"}
	}

	err := db.Preload("Amount").Preload("CartItems").Preload("CartItems.Product").Preload("CartItems.Product.PaymentSubject").Preload("CartItems.Amount").Preload("Manager").Preload("WebSite").Preload("OrderChannel").First(order, order.Id).Error
	if err != nil {
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

		err := db.Model(&Order{}).Preload("Amount").Preload("CartItems").Preload("CartItems.Product").Preload("CartItems.Amount").Preload("Manager").Preload("WebSite").Preload("OrderChannel").
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

		err := db.Model(&Order{}).Limit(limit).Preload("CartItems").Preload("CartItems.Product").Preload("CartItems.Amount").Preload("Amount").Preload("Manager").Preload("WebSite").Preload("OrderChannel").
			Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
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

	return db.Set("gorm:association_autoupdate", false).
		Model(order).Preload("Amount").Preload("CartItems").Preload("CartItems.Product").Preload("CartItems.Amount").Preload("Manager").Preload("WebSite").Preload("OrderChannel").Omit("id", "account_id").Updates(input).Error
}

func (order Order) delete () error {

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

func (order *Order) RetailPriceCalculation () error {

	/*sum := float64(0)

	if len(order.CartItems) < 1 {
		return nil
	}

	var arrProducts = make([]uint, 0)
	for _,v := range order.CartItems {
		arrProducts = append(arrProducts, v.Id)
	}

	err := db.Table("products").Where("account_id = ? AND id IN (?)", order.AccountId, arrProducts).Select("sum(retail_price)").Row().Scan(&sum)
	if err != nil {
		log.Fatal(err)
		return err
	}*/

	// fmt.Println(sum)
	// order.Cart.Value = sum
	// order.Cart.Count = len(order.CartItems)

	return nil
}
