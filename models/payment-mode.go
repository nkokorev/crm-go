package models

import (
	"errors"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)

// Признак способа расчета
type PaymentMode struct {
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	AccountId 	uint	`json:"account_id" gorm:"index;not null"` // аккаунт-владелец ключа

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

	db.Migrator().CreateTable(&PaymentMode{})

	// db.Model(&PaymentMode{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE payment_modes ADD CONSTRAINT payment_modes_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

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
func (paymentMode *PaymentMode) BeforeCreate(tx *gorm.DB) error {
	paymentMode.Id = 0
	return nil
}
func (paymentMode *PaymentMode) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(paymentMode)
	} else {
		_db = _db.Model(&PaymentMode{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{""})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}

// ######### CRUD Functions ############
func (paymentMode PaymentMode) create() (Entity, error)  {
	_item := paymentMode
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}
func (PaymentMode) get(id uint, preloads []string) (Entity, error) {

	var item PaymentMode

	err := item.GetPreloadDb(false, false, preloads).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
func (paymentMode *PaymentMode) load(preloads []string) error {
	if paymentMode.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить CartItem - не указан  Id"}
	}

	err := paymentMode.GetPreloadDb(false, false, preloads).First(paymentMode, paymentMode.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (*PaymentMode) loadByPublicId(preloads []string) error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}
func (PaymentMode) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return PaymentMode{}.getPaginationList(accountId, 0,100,sortBy,"",nil,preload)
}
func (PaymentMode) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	paymentModes := make([]PaymentMode,0)
	var total int64

	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&PaymentMode{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where("account_id IN (?)", []uint{1, accountId}).
			Find(&paymentModes, "name ILIKE ? OR code ILIKE ?", search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&PaymentMode{}).GetPreloadDb(false, false, nil).
			Where("account_id IN (?) AND name ILIKE ? OR code ILIKE ?", []uint{1, accountId}, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		err := (&PaymentMode{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where("account_id IN (?)", []uint{1, accountId}).
			Find(&paymentModes).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&PaymentMode{}).GetPreloadDb(false, false, nil).Where("account_id IN (?)", []uint{1, accountId}).Count(&total).Error
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
func (paymentMode *PaymentMode) update(input map[string]interface{}, preloads []string) error {
	// delete(input,"order")
	utils.FixInputHiddenVars(&input)
	/*if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id"}); err != nil {
		return err
	}*/

	if err := paymentMode.GetPreloadDb(false, false, nil).Where("id = ?", paymentMode.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	err := paymentMode.GetPreloadDb(false,false, preloads).First(paymentMode, paymentMode.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (paymentMode *PaymentMode) delete () error {
	return paymentMode.GetPreloadDb(true,false,nil).Where("id = ?", paymentMode.Id).Delete(paymentMode).Error
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