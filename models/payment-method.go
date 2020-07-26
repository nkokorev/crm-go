package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"log"
)

// Системные методы оплаты
type PaymentMethod struct {
	
	Id     		uint   	`json:"id" gorm:"primary_key"`
	AccountId 	uint	`json:"accountId" gorm:"index;not null"` // аккаунт-владелец ключа

	// [cash, online, credit] // yandex: bank_card, apple_pay, google_pay, yandex_money, ...]
	Code	string	`json:"code" gorm:"type:varchar(32);unique;not null;"`

	// Просто имя, выводится в магазине
	Name	string	`json:"name" gorm:"type:varchar(128);not null;"`

	WebSites	[]WebSite `json:"webSites" gorm:"many2many:payment_methods_web_sites;preload"`
	Payments	[]Payment `json:"payments" gorm:"many2many:payment_methods_payments;preload"`
	// Доступен ли данный способ платежей
	// Enabled	bool	`json:"enabled" gorm:"type:bool;default:false"`
}

// ############# Entity interface #############
func (paymentMethod PaymentMethod) GetId() uint { return paymentMethod.Id }
func (paymentMethod *PaymentMethod) setId(id uint) { paymentMethod.Id = id }
func (paymentMethod PaymentMethod) GetAccountId() uint { return paymentMethod.AccountId }
func (paymentMethod *PaymentMethod) setAccountId(id uint) { paymentMethod.AccountId = id }
func (paymentMethod PaymentMethod) SystemEntity() bool { return paymentMethod.AccountId == 1 }

// ############# Entity interface #############

func (PaymentMethod) PgSqlCreate() {
	db.CreateTable(&PaymentMethod{})
	db.Model(&PaymentMethod{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")

	paymentMethods := []PaymentMethod {
		{Name:   "Оплата при получении",	Code: "cash"},
		{Name:   "Онлайн-оплата картой",	Code: "online"},
		{Name:   "Покупка в кредит",		Code: "credit"},
	}

	for i := range(paymentMethods) {
		_, err := Account{Id: 1}.CreateEntity(&paymentMethods[i])
		if err != nil {
			log.Fatalf("Не удалось создать paymentMethods: ", err)
		}
	}
	
}
func (paymentMethod *PaymentMethod) BeforeCreate(scope *gorm.Scope) error {
	paymentMethod.Id = 0
	return nil
}

// ######### CRUD Functions ############
func (paymentMethod PaymentMethod) create() (Entity, error)  {
	_paymentMethod := paymentMethod
	if err := db.Create(&_paymentMethod).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_paymentMethod

	return entity, nil
}

func (PaymentMethod) get(id uint) (Entity, error) {

	var paymentMethod PaymentMethod

	err := db.First(&paymentMethod, id).Error
	if err != nil {
		return nil, err
	}
	return &paymentMethod, nil
}
func (paymentMethod *PaymentMethod) load() error {
	if paymentMethod.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить PaymentMethod - не указан  Id"}
	}

	err := db.First(paymentMethod, paymentMethod.Id).Error
	if err != nil {
		return err
	}
	return nil
}

func (PaymentMethod) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return PaymentMethod{}.getPaginationList(accountId, 0,100,sortBy,"")
}

func (PaymentMethod) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	paymentMethods := make([]PaymentMethod,0)
	var total uint

	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&PaymentMethod{}).Limit(limit).Offset(offset).Order(sortBy).Where("account_id IN (?)", []uint{1, accountId}).
			Find(&paymentMethods, "name ILIKE ? OR code ILIKE ?", search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&PaymentMethod{}).
			Where("account_id IN (?) AND name ILIKE ? OR code ILIKE ?", []uint{1, accountId}, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		err := db.Model(&PaymentMethod{}).Limit(limit).Offset(offset).Order(sortBy).Where("account_id IN (?)", []uint{1, accountId}).
			Find(&paymentMethods).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&PaymentMethod{}).Where("account_id IN (?)", []uint{1, accountId}).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	entities := make([]Entity,len(paymentMethods))
	for i,_ := range paymentMethods {
		entities[i] = &paymentMethods[i]
	}

	return entities, total, nil
}

func (paymentMethod *PaymentMethod) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).
		Model(paymentMethod).Omit("id", "account_id").Updates(input).Error
}

func (paymentMethod PaymentMethod) delete () error {
	return db.Model(PaymentMethod{}).Where("id = ?", paymentMethod.Id).Delete(paymentMethod).Error
}
// ######### END CRUD Functions ############


// ########## Work function ############
func (PaymentMethod) getByCode(accountId uint, code string) (*PaymentMethod, error) {
	var paymentMethod PaymentMethod
	fmt.Println(accountId, code)
	err := db.First(&paymentMethod,"account_id IN (?) AND code = ?", []uint{1, accountId}, code).Error
	return &paymentMethod, err
}
