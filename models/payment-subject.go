package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"log"
)

// Признак предмета расчета
type PaymentSubject struct {
	Id     		uint   	`json:"id" gorm:"primary_key"`
	AccountId 	uint	`json:"accountId" gorm:"index;not null"` // аккаунт-владелец ключа

	Name	string	`json:"name" gorm:"type:varchar(128);unique;not null;"`
	Code	string	`json:"code" gorm:"type:varchar(32);unique;not null;"`
}

// ############# Entity interface #############
func (paymentSubject PaymentSubject) GetId() uint { return paymentSubject.Id }
func (paymentSubject *PaymentSubject) setId(id uint) { paymentSubject.Id = id }
func (paymentSubject PaymentSubject) GetAccountId() uint { return paymentSubject.AccountId }
func (paymentSubject *PaymentSubject) setAccountId(id uint) { paymentSubject.AccountId = id }
func (paymentSubject PaymentSubject) SystemEntity() bool { return paymentSubject.AccountId == 1 }

// ############# Entity interface #############

func (PaymentSubject) PgSqlCreate() {
	db.CreateTable(&PaymentSubject{})
	db.Model(&PaymentSubject{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")

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
func (paymentSubject *PaymentSubject) BeforeCreate(scope *gorm.Scope) error {
	paymentSubject.Id = 0
	return nil
}

// ######### CRUD Functions ############
func (paymentSubject PaymentSubject) create() (Entity, error)  {
	_productSubject := paymentSubject
	if err := db.Create(&_productSubject).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_productSubject

	return entity, nil
}

func (PaymentSubject) get(id uint) (Entity, error) {

	var paymentSubject PaymentSubject

	err := db.First(&paymentSubject, id).Error
	if err != nil {
		return nil, err
	}
	return &paymentSubject, nil
}
func (paymentSubject *PaymentSubject) load() error {
	if paymentSubject.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить PaymentSubject - не указан  Id"}
	}

	err := db.First(paymentSubject, paymentSubject.Id).Error
	if err != nil {
		return err
	}
	return nil
}

func (PaymentSubject) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return PaymentSubject{}.getPaginationList(accountId, 0,100,sortBy,"")
}

func (PaymentSubject) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	paymentSubjects := make([]PaymentSubject,0)
	var total uint

	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&PaymentSubject{}).Limit(limit).Offset(offset).Order(sortBy).Where("account_id IN (?)", []uint{1, accountId}).
			Find(&paymentSubjects, "name ILIKE ? OR code ILIKE ?", search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&PaymentSubject{}).
			Where("account_id IN (?) AND name ILIKE ? OR code ILIKE ?", []uint{1, accountId}, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		err := db.Model(&PaymentSubject{}).Limit(limit).Offset(offset).Order(sortBy).Where("account_id IN (?)", []uint{1, accountId}).
			Find(&paymentSubjects).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&PaymentSubject{}).Where("account_id IN (?)", []uint{1, accountId}).Count(&total).Error
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

func (paymentSubject *PaymentSubject) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).
		Model(paymentSubject).Omit("id", "account_id").Updates(input).Error
}

func (paymentSubject *PaymentSubject) delete () error {
	return db.Model(PaymentSubject{}).Where("id = ?", paymentSubject.Id).Delete(paymentSubject).Error
}
// ######### END CRUD Functions ############


// ########## Work function ############

