package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
)

//Объект платежа - кто-то, что-то вам заплатил. Или хочет заплатить. Или должен...

type PaymentAmount struct {
	
	Id     		uint   	`json:"id" gorm:"primary_key"`
	AccountId 	uint	`json:"accountId" gorm:"index;not null"` // аккаунт-владелец ключа

	// Сумма расчета
	Value 	float64	`json:"value" gorm:"type:numeric(15,2);default:0"`

	// КОД валюты в  ISO-4217 https://www.iso.org/iso-4217-currency-codes.html
	Currency 	string 	`json:"currency" gorm:"type:varchar(3);default:'RUB'"`
}

// ############# Entity interface #############
func (paymentAmount PaymentAmount) GetId() uint { return paymentAmount.Id }
func (paymentAmount *PaymentAmount) setId(id uint) { paymentAmount.Id = id }
func (paymentAmount *PaymentAmount) setPublicId(id uint) { }
func (paymentAmount PaymentAmount) GetAccountId() uint { return paymentAmount.AccountId }
func (paymentAmount *PaymentAmount) setAccountId(id uint) { paymentAmount.AccountId = id }
func (paymentAmount PaymentAmount) SystemEntity() bool { return false }

// ############# Entity interface #############

func (PaymentAmount) PgSqlCreate() {
	db.AutoMigrate(&PaymentAmount{})
	db.Model(&PaymentAmount{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}
func (paymentAmount *PaymentAmount) BeforeCreate(scope *gorm.Scope) error {
	paymentAmount.Id = 0
	return nil
}

// ######### CRUD Functions ############
func (paymentAmount PaymentAmount) create() (Entity, error)  {
	_paymentAmount := paymentAmount
	if err := db.Create(&_paymentAmount).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_paymentAmount

	return entity, nil
}

func (PaymentAmount) get(id uint) (Entity, error) {

	var paymentAmount PaymentAmount

	err := db.First(&paymentAmount, id).Error
	if err != nil {
		return nil, err
	}
	return &paymentAmount, nil
}
func (paymentAmount *PaymentAmount) load() error {
	if paymentAmount.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить PaymentAmount - не указан  Id"}
	}

	err := db.First(paymentAmount, paymentAmount.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (*PaymentAmount) loadByPublicId() error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}
func (PaymentAmount) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return PaymentAmount{}.getPaginationList(accountId, 0,100,sortBy,"")
}

func (PaymentAmount) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	paymentSubjects := make([]PaymentAmount,0)
	var total uint

	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&PaymentAmount{}).Limit(limit).Offset(offset).Order(sortBy).Where("account_id  = ?", accountId).
			Find(&paymentSubjects, "value ILIKE ? OR currency ILIKE ?", search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&PaymentAmount{}).
			Where("account_id = ? AND value ILIKE ? OR currency ILIKE ?", accountId, search, search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		err := db.Model(&PaymentAmount{}).Limit(limit).Offset(offset).Order(sortBy).Where("account_id = ?", accountId).
			Find(&paymentSubjects).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&PaymentAmount{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	entities := make([]Entity,len(paymentSubjects))
	for i,_ := range paymentSubjects {
		entities[i] = &paymentSubjects[i]
	}

	return entities, total, nil
}

func (paymentAmount *PaymentAmount) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).
		Model(paymentAmount).Omit("id", "account_id").Updates(input).Error
}

func (paymentAmount *PaymentAmount) delete () error {
	return db.Where("id = ?", paymentAmount.Id).Delete(paymentAmount).Error
}
func (PaymentAmount) deletes (paymentsIds []uint) error {
	return db.Where("id IN (?)", paymentsIds).Delete(&PaymentAmount{}).Error
}
// ######### END CRUD Functions ############


// ########## Work function ############

