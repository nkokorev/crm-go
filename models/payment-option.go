package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
)

// Методы оплаты        payment option,
type PaymentOption struct {
	
	Id     		uint   	`json:"id" gorm:"primary_key"`
	AccountId 	uint	`json:"accountId" gorm:"index;not null"` // аккаунт-владелец ключа

	// [cash, online, credit] // yandex: bank_card, apple_pay, google_pay, yandex_money, ...]
	Code	string	`json:"code" gorm:"type:varchar(32);unique;not null;"`

	// Просто имя, выводится в магазине
	Name	string	`json:"name" gorm:"type:varchar(128);not null;"`

	// Связывает с реальным способом оплаты
	OwnerId   	uint	//1,2 ...
	OwnerType	string	// payment_yandex, payment_chash

	WebSites	[]WebSite `json:"webSites" gorm:"many2many:payment_options_web_sites;preload"`
	Payments	[]Payment `json:"payments" gorm:"many2many:payment_options_payments;preload"`

	DeliveryPickUps 	[]DeliveryPickup  `json:"deliveryPickUps" gorm:"many2many:payment_options_delivery_pickups;preload"`
	DeliveryCouriers 	[]DeliveryCourier  `json:"deliveryCouriers" gorm:"many2many:payment_options_delivery_couriers;preload"`
	DeliveryRussianPosts 	[]DeliveryRussianPost  `json:"deliveryRussianPosts" gorm:"many2many:payment_options_delivery_russian_posts;preload"`

	PaymentMethod	`json:"-" gorm:"-"`
	// Доступен ли данный способ платежей
	// Enabled	bool	`json:"enabled" gorm:"type:bool;default:false"`
}

// ############# Entity interface #############
func (paymentOption PaymentOption) GetId() uint { return paymentOption.Id }
func (paymentOption *PaymentOption) setId(id uint) { paymentOption.Id = id }
func (paymentOption PaymentOption) GetAccountId() uint { return paymentOption.AccountId }
func (paymentOption *PaymentOption) setAccountId(id uint) { paymentOption.AccountId = id }
func (paymentOption PaymentOption) SystemEntity() bool { return false }

// ############# Entity interface #############

func (PaymentOption) PgSqlCreate() {
	db.CreateTable(&PaymentOption{})
	db.Model(&PaymentOption{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	
}
func (paymentOption *PaymentOption) BeforeCreate(scope *gorm.Scope) error {
	paymentOption.Id = 0
	return nil
}
func (paymentOption *PaymentOption) AfterFind() (err error) {

	switch paymentOption.OwnerType {
	case "payment_yandexes":
		var paymentYandex PaymentYandex
		if err := db.First(&paymentYandex, paymentOption.OwnerId).Error; err != nil { return err}
		paymentOption.PaymentMethod = &paymentYandex
	case "payment_cashes":
		var paymentCash PaymentCash
		if err := db.First(&paymentCash, paymentOption.OwnerId).Error; err != nil { return err}
		paymentOption.PaymentMethod = &paymentCash
	}

	return nil
}

// ######### CRUD Functions ############
func (paymentOption PaymentOption) create() (Entity, error)  {
	_paymentMethod := paymentOption
	if err := db.Create(&_paymentMethod).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_paymentMethod

	return entity, nil
}

func (PaymentOption) get(id uint) (Entity, error) {

	var paymentOption PaymentOption

	err := db.First(&paymentOption, id).Error
	if err != nil {
		return nil, err
	}
	return &paymentOption, nil
}
func (paymentOption *PaymentOption) load() error {
	if paymentOption.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить PaymentOption - не указан  Id"}
	}

	err := db.First(paymentOption, paymentOption.Id).Error
	if err != nil {
		return err
	}
	return nil
}

func (PaymentOption) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return PaymentOption{}.getPaginationList(accountId, 0,100,sortBy,"")
}

func (PaymentOption) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	paymentOptions := make([]PaymentOption,0)
	var total uint

	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&PaymentOption{}).Limit(limit).Offset(offset).Order(sortBy).Where("account_id = ?", accountId).
			Find(&paymentOptions, "name ILIKE ? OR code ILIKE ?", search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&PaymentOption{}).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ?", accountId, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		err := db.Model(&PaymentOption{}).Limit(limit).Offset(offset).Order(sortBy).Where("account_id = ?", accountId).
			Find(&paymentOptions).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&PaymentOption{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	entities := make([]Entity,len(paymentOptions))
	for i,_ := range paymentOptions {
		entities[i] = &paymentOptions[i]
	}

	return entities, total, nil
}

func (paymentOption *PaymentOption) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).
		Model(paymentOption).Omit("id", "account_id").Updates(input).Error
}

func (paymentOption PaymentOption) delete () error {
	return db.Model(PaymentOption{}).Where("id = ?", paymentOption.Id).Delete(paymentOption).Error
}
// ######### END CRUD Functions ############


// ########## Work function ############
func (PaymentOption) getByCode(accountId uint, code string) (*PaymentOption, error) {
	var paymentOption PaymentOption
	fmt.Println(accountId, code)
	err := db.First(&paymentOption,"account_id IN (?) AND code = ?", []uint{1, accountId}, code).Error
	return &paymentOption, err
}

func (paymentOption PaymentOption) AppendDeliveryCourier(deliveryCourier DeliveryCourier) error {
	if err := db.Model(&paymentOption).Association("DeliveryCouriers").Append(deliveryCourier).Error; err != nil {
		return err
	}

	return nil
}

