package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"log"
)

// Признак способа расчета
type PaymentMode struct {
	Id     		uint   	`json:"id" gorm:"primary_key"`
	AccountId 	uint	`json:"accountId" gorm:"index;not null"` // аккаунт-владелец ключа

	Name	string	`json:"name" gorm:"type:varchar(128);unique;not null;"`
	Code	string	`json:"code" gorm:"type:varchar(32);unique;not null;"`
}

// ############# Entity interface #############
func (paymentMode PaymentMode) GetId() uint { return paymentMode.Id }
func (paymentMode *PaymentMode) setId(id uint) { paymentMode.Id = id }
func (paymentMode *PaymentMode) setPublicId(id uint) { paymentMode.Id = id }
func (paymentMode PaymentMode) GetAccountId() uint { return paymentMode.AccountId }
func (paymentMode *PaymentMode) setAccountId(id uint) { paymentMode.AccountId = id }
func (paymentMode PaymentMode) SystemEntity() bool { return paymentMode.AccountId == 1 }

// ############# Entity interface #############

func (PaymentMode) PgSqlCreate() {

	db.AutoMigrate(&PaymentMode{})

	db.Model(&PaymentMode{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")

	paymentModes := []PaymentMode {
		{Name:   "Полная предоплата",		Code: "full_prepayment"},
		{Name:   "Частичная предоплата",	Code: "partial_prepayment"},
		{Name:   "Аванс",					Code: "advance"},
		{Name:   "Услуга",					Code: "service"},
		{Name:   "Полный расчет",			Code: "full_payment"},
		{Name:   "Частичный расчет и кредит",	Code: "partial_payment"},
		{Name:   "Кредит",					Code: "credit"},
		{Name:   "Выплата по кредиту",		Code: "credit_payment"},
	}

	for i := range(paymentModes) {
		_, err := Account{Id: 1}.CreateEntity(&paymentModes[i])
		if err != nil {
			log.Fatalf("Не удалось создать paymentModes: ", err)
		}
	}
	
}
func (paymentMode *PaymentMode) BeforeCreate(scope *gorm.Scope) error {
	paymentMode.Id = 0
	return nil
}

// ######### CRUD Functions ############
func (paymentMode PaymentMode) create() (Entity, error)  {
	_productSubject := paymentMode
	if err := db.Create(&_productSubject).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_productSubject

	return entity, nil
}
func (PaymentMode) get(id uint) (Entity, error) {

	var paymentMode PaymentMode

	err := db.First(&paymentMode, id).Error
	if err != nil {
		return nil, err
	}
	return &paymentMode, nil
}
func (paymentMode *PaymentMode) load() error {
	if paymentMode.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить PaymentMode - не указан  Id"}
	}

	err := db.First(paymentMode, paymentMode.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (*PaymentMode) loadByPublicId() error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}
func (PaymentMode) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return PaymentMode{}.getPaginationList(accountId, 0,100,sortBy,"",nil)
}
func (PaymentMode) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{}) ([]Entity, uint, error) {

	paymentModes := make([]PaymentMode,0)
	var total uint

	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&PaymentMode{}).Limit(limit).Offset(offset).Order(sortBy).Where("account_id IN (?)", []uint{1, accountId}).
			Find(&paymentModes, "name ILIKE ? OR code ILIKE ?", search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&PaymentMode{}).
			Where("account_id IN (?) AND name ILIKE ? OR code ILIKE ?", []uint{1, accountId}, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		err := db.Model(&PaymentMode{}).Limit(limit).Offset(offset).Order(sortBy).Where("account_id IN (?)", []uint{1, accountId}).
			Find(&paymentModes).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&PaymentMode{}).Where("account_id IN (?)", []uint{1, accountId}).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	entities := make([]Entity,len(paymentModes))
	for i := range paymentModes {
		entities[i] = &paymentModes[i]
	}

	return entities, total, nil
}
func (paymentMode *PaymentMode) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).
		Model(paymentMode).Omit("id", "account_id").Updates(input).Error
}
func (paymentMode *PaymentMode) delete () error {
	return db.Model(PaymentMode{}).Where("id = ?", paymentMode.Id).Delete(paymentMode).Error
}
// ######### END CRUD Functions ############

func (PaymentMode) GetFullPrepaymentMode() (PaymentMode, error) {
	var mode PaymentMode

	err := db.First(&mode, "code = 'full_prepayment'").Error
	if err != nil {
		return mode, err
	}
	return mode, nil
}
func (PaymentMode) GetPartialPrepaymentMode() (PaymentMode, error) {
	var mode PaymentMode

	err := db.First(&mode, "code = 'partial_prepayment'").Error
	if err != nil {
		return mode, err
	}
	return mode, nil
}
func (PaymentMode) GetFullPaymentMode() (PaymentMode, error) {
	var mode PaymentMode

	err := db.First(&mode, "code = 'full_payment'").Error
	if err != nil {
		return mode, err
	}
	return mode, nil
}