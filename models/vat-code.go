package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"log"
)

// Признак предмета расчета
type VatCode struct {
	
	Id     		uint   	`json:"id" gorm:"primary_key"`
	AccountId 	uint	`json:"accountId" gorm:"index;not null"` // аккаунт-владелец ключа

	Name	string	`json:"name" gorm:"type:varchar(128);unique;not null;"`
	Code	string	`json:"code" gorm:"type:varchar(32);unique;not null;"`

	// Системный ID у яндекса, подробнее: https://kassa.yandex.ru/developers/54fz/parameters-values#vat-codes
	YandexCode	uint	`json:"yandexCode" gorm:"type:int;unique;not null;"`
}

// ############# Entity interface #############
func (vatCode VatCode) GetId() uint { return vatCode.Id }
func (vatCode *VatCode) setId(id uint) { vatCode.Id = id }
func (vatCode VatCode) GetAccountId() uint { return vatCode.AccountId }
func (vatCode *VatCode) setAccountId(id uint) { vatCode.AccountId = id }
func (vatCode VatCode) SystemEntity() bool { return vatCode.AccountId == 1 }

// ############# Entity interface #############

func (VatCode) PgSqlCreate() {
	db.CreateTable(&VatCode{})
	db.Model(&VatCode{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")

	vatCodes := []VatCode {
		{Name:   "Без НДС",				Code: "without", YandexCode: 1},
		{Name:   "НДС по ставке 0%",	Code: "rate_0", YandexCode: 2},
		{Name:   "НДС по ставке 10%",	Code: "rate_10", YandexCode: 3},
		{Name:   "НДС чека по ставке 20%",	Code: "receipt_rate_20", YandexCode: 4},
		{Name:   "НДС чека по расчетной ставке 10/110",	Code: "receipt_estimated_rate_10/110", YandexCode: 5},
		{Name:   "НДС чека по расчетной ставке 20/120",	Code: "receipt_estimated_rate_20/120", YandexCode: 6},
	}

	for i := range(vatCodes) {
		_, err := Account{Id: 1}.CreateEntity(&vatCodes[i])
		if err != nil {
			log.Fatalf("Не удалось создать vatCodes: ", err)
		}
	}
	
}
func (vatCode *VatCode) BeforeCreate(scope *gorm.Scope) error {
	vatCode.Id = 0
	return nil
}

// ######### CRUD Functions ############
func (vatCode VatCode) create() (Entity, error)  {
	_productSubject := vatCode
	if err := db.Create(&_productSubject).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_productSubject

	return entity, nil
}

func (VatCode) get(id uint) (Entity, error) {

	var vatCode VatCode

	err := db.First(&vatCode, id).Error
	if err != nil {
		return nil, err
	}
	return &vatCode, nil
}
func (vatCode *VatCode) load() error {
	if vatCode.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить VatCode - не указан  Id"}
	}

	err := db.First(vatCode, vatCode.Id).Error
	if err != nil {
		return err
	}
	return nil
}

func (VatCode) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return VatCode{}.getPaginationList(accountId, 0,100,sortBy,"")
}

func (VatCode) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	vatCodes := make([]VatCode,0)
	var total uint

	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&VatCode{}).Limit(limit).Offset(offset).Order(sortBy).Where("account_id IN (?)", []uint{1, accountId}).
			Find(&vatCodes, "name ILIKE ? OR code ILIKE ?", search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&VatCode{}).
			Where("account_id IN (?) AND name ILIKE ? OR code ILIKE ?", []uint{1, accountId}, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		err := db.Model(&VatCode{}).Limit(limit).Offset(offset).Order(sortBy).Where("account_id IN (?)", []uint{1, accountId}).
			Find(&vatCodes).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&VatCode{}).Where("account_id IN (?)", []uint{1, accountId}).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	entities := make([]Entity,len(vatCodes))
	for i,_ := range vatCodes {
		entities[i] = &vatCodes[i]
	}

	return entities, total, nil
}

func (vatCode *VatCode) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).
		Model(vatCode).Omit("id", "account_id").Updates(input).Error
}

func (vatCode *VatCode) delete () error {
	return db.Model(VatCode{}).Where("id = ?", vatCode.Id).Delete(vatCode).Error
}
// ######### END CRUD Functions ############


// ########## Work function ############
