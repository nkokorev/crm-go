package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nkokorev/crm-go/utils"
	"strings"
)

// Условный метод оплаты кешем. Это либо нал, либо перевод на карту.
type PaymentCash struct {
	Id     		uint   	`json:"id" gorm:"primary_key"`
	HashId 		string `json:"hashId" gorm:"type:varchar(12);unique_index;not null;"` // публичный Id для защиты от спама/парсинга
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	Name 		string 	`json:"name" gorm:"type:varchar(128);default:''"` // Имя интеграции магазина "<name>"
	Description 		string 	`json:"description" gorm:"type:varchar(255);default:''"` // Описание метода оплаты

	// Включен ли данный способ оплаты ??
	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"`

	PaymentOption   PaymentOption `gorm:"polymorphic:Owner;"`
}

func (PaymentCash) PgSqlCreate() {
	db.CreateTable(&PaymentCash{})
	db.Model(&PaymentCash{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}
func (paymentCash *PaymentCash) BeforeCreate(scope *gorm.Scope) error {
	paymentCash.Id = 0
	paymentCash.HashId = strings.ToLower(utils.RandStringBytesMaskImprSrcUnsafe(12, true))
	return nil
}
// ############# Entity interface #############
func (paymentCash PaymentCash) GetId() uint { return paymentCash.Id }
func (paymentCash *PaymentCash) setId(id uint) { paymentCash.Id = id }
func (paymentCash PaymentCash) GetAccountId() uint { return paymentCash.AccountId }
func (paymentCash *PaymentCash) setAccountId(id uint) { paymentCash.AccountId = id }
func (PaymentCash) SystemEntity() bool { return false }

// ############# Entity interface #############


// ######### CRUD Functions ############
func (paymentCash PaymentCash) create() (Entity, error)  {
	wb := paymentCash
	if err := db.Create(&wb).Error; err != nil {
		return nil, err
	}

	var entity Entity = &wb

	return entity, nil
}

func (PaymentCash) get(id uint) (Entity, error) {

	var paymentCash PaymentCash

	err := db.First(&paymentCash, id).Error
	if err != nil {
		return nil, err
	}
	return &paymentCash, nil
}
func (paymentCash *PaymentCash) load() error {
	if paymentCash.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить PaymentCash - не указан  Id"}
	}

	err := db.First(paymentCash,paymentCash.Id).Error
	if err != nil {
		return err
	}
	return nil
}

func (PaymentCash) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return  PaymentCash{}.getPaginationList(accountId, 0, 100, sortBy, "")
}

func (PaymentCash) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	paymentCashs := make([]PaymentCash,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&PaymentCash{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&paymentCashs, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&PaymentCash{}).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&PaymentCash{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&paymentCashs).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&PaymentCash{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(paymentCashs))
	for i,_ := range paymentCashs {
		entities[i] = &paymentCashs[i]
	}

	return entities, total, nil
}

func (PaymentCash) getByEvent(eventName string) (*PaymentCash, error) {

	wh := PaymentCash{}

	if err := db.First(&wh, "event_type = ?", eventName).Error; err != nil {
		return nil, err
	}

	return &wh, nil
}

func (paymentCash *PaymentCash) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).
		Model(paymentCash).Where("id", paymentCash.Id).Omit("id", "account_id").Updates(input).Error
}

func (paymentCash *PaymentCash) delete () error {
	return db.Model(PaymentCash{}).Where("id = ?", paymentCash.Id).Delete(paymentCash).Error
}
// ######### END CRUD Functions ############


// ########## Work function ############

func (paymentCash PaymentCash) CreatePayment(order Order) (*Payment, error) {

	_p := Payment {
		AccountId: paymentCash.AccountId,
		Paid: false,
		Amount: order.Amount,
		IncomeAmount: order.Amount,
		RefundedAmount: PaymentAmount{AccountId: order.AccountId, Value: float64(0), Currency: "RUB"},
		Description:  fmt.Sprintf("Заказ №%v в магазине AiroCliamte", order.Id),  // Видит клиент
		PaymentMethodData: PaymentMethodData{Type: "bank_card"}, // вообще еще вопрос

		// Чтобы понять какой платеж был оплачен!!!
		Metadata: postgres.Jsonb{ RawMessage: utils.MapToRawJson(map[string]interface{}{
			"orderId":order.Id,
			"accountId":paymentCash.AccountId,
		})},
		SavePaymentMethod: false,
		OwnerId: paymentCash.Id,
		OwnerType: "payment_cashes",
		OrderId: order.Id,
	}

	// создаем внутри платеж
	entity, err := _p.create()
	if err != nil {
		return nil, err
	}
	payment := entity.(*Payment)
	
	return payment, nil
}


func (paymentCash PaymentCash) SetPaymentOption(paymentOption PaymentOption) error {
	if err := db.Model(&paymentCash).Association("PaymentOption").Append(paymentOption).Error; err != nil {
		return err
	}

	return nil
}