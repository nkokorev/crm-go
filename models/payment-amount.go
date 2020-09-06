package models

import (
	"errors"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)

//Объект платежа - кто-то, что-то вам заплатил. Или хочет заплатить. Или должен...

type PaymentAmount struct {
	
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	AccountId 	uint	`json:"account_id" gorm:"index;not null"` // аккаунт-владелец ключа

	// Сумма расчета
	// Value 	float64	`json:"value" gorm:"type:numeric(15,2);"`
	Value 		float64	`json:"value" gorm:"type:numeric;"`

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
	if err := db.Migrator().CreateTable(&PaymentAmount{}); err != nil {
		log.Fatal(err)
	}
	// db.Model(&PaymentAmount{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE payment_amounts ADD CONSTRAINT payment_amounts_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (paymentAmount *PaymentAmount) BeforeCreate(tx *gorm.DB) error {
	paymentAmount.Id = 0
	return nil
}
func (paymentAmount *PaymentAmount) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(&paymentAmount)
	} else {
		_db = _db.Model(&PaymentAmount{})
	}

	if autoPreload {
		return db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{""})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}

// ######### CRUD Functions ############
func (paymentAmount PaymentAmount) create() (Entity, error)  {
	_item := paymentAmount
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,true, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}

func (PaymentAmount) get(id uint, preloads []string) (Entity, error) {

	var item PaymentAmount

	err := item.GetPreloadDb(false, false, preloads).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
func (paymentAmount *PaymentAmount) load(preloads []string) error {
	if paymentAmount.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить CartItem - не указан  Id"}
	}

	err := paymentAmount.GetPreloadDb(false, false, preloads).First(paymentAmount, paymentAmount.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (*PaymentAmount) loadByPublicId(preloads []string) error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}
func (PaymentAmount) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return PaymentAmount{}.getPaginationList(accountId, 0,25,sortBy,"",nil,preload)
}

func (PaymentAmount) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	paymentSubjects := make([]PaymentAmount,0)
	var total int64

	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&PaymentAmount{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where("account_id  = ?", accountId).
			Find(&paymentSubjects, "value ILIKE ? OR currency ILIKE ?", search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&PaymentAmount{}).GetPreloadDb(false, false, nil).
			Where("account_id = ? AND value ILIKE ? OR currency ILIKE ?", accountId, search, search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		err := (&PaymentAmount{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where("account_id = ?", accountId).
			Find(&paymentSubjects).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&PaymentAmount{}).GetPreloadDb(false, false, nil).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	entities := make([]Entity,len(paymentSubjects))
	for i := range paymentSubjects {
		entities[i] = &paymentSubjects[i]
	}

	return entities, total, nil
}

func (paymentAmount *PaymentAmount) update(input map[string]interface{}, preloads []string) error {
	// delete(input,"order")
	utils.FixInputHiddenVars(&input)
	/*if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id"}); err != nil {
		return err
	}*/

	if err := paymentAmount.GetPreloadDb(false, false, nil).Where("id = ?", paymentAmount.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	err := paymentAmount.GetPreloadDb(false,false, preloads).First(paymentAmount, paymentAmount.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (paymentAmount *PaymentAmount) delete () error {
	return paymentAmount.GetPreloadDb(true,false,nil).Where("id = ?", paymentAmount.Id).Delete(paymentAmount).Error
}
func (PaymentAmount) deletes (paymentsIds []uint) error {
	return db.Where("id IN (?)", paymentsIds).Delete(&PaymentAmount{}).Error
}
// ######### END CRUD Functions ############
