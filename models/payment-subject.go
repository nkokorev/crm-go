package models

import (
	"errors"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)

// Признак предмета расчета
type PaymentSubject struct {
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	AccountId 	uint	`json:"account_id" gorm:"index;not null"` // аккаунт-владелец ключа

	Name	string	`json:"name" gorm:"type:varchar(128);unique;not null;"`
	Code	string	`json:"code" gorm:"type:varchar(32);unique;not null;"`
}

// ############# Entity interface #############
func (paymentSubject PaymentSubject) GetId() uint { return paymentSubject.Id }
func (paymentSubject *PaymentSubject) setId(id uint) { paymentSubject.Id = id }
func (paymentSubject *PaymentSubject) setPublicId(id uint) {}
func (paymentSubject PaymentSubject) GetAccountId() uint { return paymentSubject.AccountId }
func (paymentSubject *PaymentSubject) setAccountId(id uint) { paymentSubject.AccountId = id }
func (paymentSubject PaymentSubject) SystemEntity() bool { return paymentSubject.AccountId == 1 }

// ############# Entity interface #############

func (PaymentSubject) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&PaymentSubject{}); err != nil {
		log.Fatal(err)
	}
	// db.Model(&PaymentSubject{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE payment_subjects ADD CONSTRAINT payment_subjects_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

	paymentSubjects := []PaymentSubject {
		{Name:   "Товар",		Code: "commodity"},
		{Name:   "Подакцизный товар",	Code: "excise"},
		{Name:   "Работа",			Code: "job"},
		{Name:   "Услуга",			Code: "service"},
		{Name:   "Ставка в азартной игре",	Code: "gambling_bet"},
		{Name:   "Выигрыш в азартной игре",	Code: "gambling_prize"},
		{Name:   "Лотерейный билет",	Code: "lottery"},
		{Name:   "Выигрыш в лотерею",	Code: "lottery_prize"},
		{Name:   "Результаты интеллектуальной деятельности",	Code: "intellectual_activity"},
		{Name:   "Платеж",	Code: "payment"},
		{Name:   "Агентское вознаграждение",Code: "agent_commission"},
		{Name:   "Имущественные права",Code: "property_right"},
		{Name:   "Внереализационный доход",Code: "non_operating_gain"},
		{Name:   "Страховой сбор",	Code: "insurance_premium"},
		{Name:   "Торговый сбор",	Code: "sales_tax"},
		{Name:   "Курортный сбор",	Code: "resort_fee"},
		{Name:   "Несколько вариантов",	Code: "composite"},
		{Name:   "Другое",	Code: "another"},
	}

	for i := range(paymentSubjects) {
		_, err := Account{Id: 1}.CreateEntity(&paymentSubjects[i])
		if err != nil {
			log.Fatalf("Не удалось создать paymentSubjects: ", err)
		}
	}
	
}
func (paymentSubject *PaymentSubject) BeforeCreate(tx *gorm.DB) error {
	paymentSubject.Id = 0
	return nil
}
func (paymentSubject *PaymentSubject) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(&paymentSubject)
	} else {
		_db = _db.Model(&PaymentSubject{})
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
func (paymentSubject PaymentSubject) create() (Entity, error)  {
	_item := paymentSubject
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}

func (PaymentSubject) get(id uint, preloads []string) (Entity, error) {

	var item PaymentSubject

	err := item.GetPreloadDb(false, false, preloads).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
func (paymentSubject *PaymentSubject) load(preloads []string) error {
	if paymentSubject.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить CartItem - не указан  Id"}
	}

	err := paymentSubject.GetPreloadDb(false, false, preloads).First(paymentSubject, paymentSubject.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (*PaymentSubject) loadByPublicId(preloads []string) error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}
func (PaymentSubject) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return PaymentSubject{}.getPaginationList(accountId, 0,100,sortBy,"",nil, preload)
}
func (PaymentSubject) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	paymentSubjects := make([]PaymentSubject,0)
	var total int64

	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&PaymentSubject{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where("account_id IN (?)", []uint{1, accountId}).
			Find(&paymentSubjects, "name ILIKE ? OR code ILIKE ?", search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&PaymentSubject{}).GetPreloadDb(false, false, nil).
			Where("account_id IN (?) AND name ILIKE ? OR code ILIKE ?", []uint{1, accountId}, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		err := (&PaymentSubject{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where("account_id IN (?)", []uint{1, accountId}).
			Find(&paymentSubjects).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&PaymentSubject{}).GetPreloadDb(false, false, nil).Where("account_id IN (?)", []uint{1, accountId}).Count(&total).Error
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

func (paymentSubject *PaymentSubject) update(input map[string]interface{}, preloads []string) error {
	// delete(input,"order")
	utils.FixInputHiddenVars(&input)
	/*if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id"}); err != nil {
		return err
	}*/

	if err := paymentSubject.GetPreloadDb(false, false, nil).Where("id = ?", paymentSubject.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	err := paymentSubject.GetPreloadDb(false,false, preloads).First(paymentSubject, paymentSubject.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (paymentSubject *PaymentSubject) delete () error {
	return paymentSubject.GetPreloadDb(true,false,nil).Where("id = ?", paymentSubject.Id).Delete(paymentSubject).Error
}
// ######### END CRUD Functions ############



