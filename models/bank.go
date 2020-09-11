package models

import (
	"fmt"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

// Расчетный счет
type Bank struct {
	Id     		uint	`json:"id" gorm:"primaryKey"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;"`
	
	// Наименование: АО "АЛЬФА-БАНК"
	Name 		string `json:"code" gorm:"type:varchar(255);"`
	Address		string `json:"address" gorm:"type:varchar(255);"`

	// БИК
	RCBIC		string `json:"rcbic" gorm:"type:varchar(9);"`
	SWIFT		string `json:"swift" gorm:"type:varchar(11);"`
	INN			string `json:"inn" gorm:"type:varchar(12);"`
	KPP			string `json:"kpp" gorm:"type:varchar(9);"`

	// Корсчет
	CorporateAccount	string `json:"corporate_account" gorm:"type:varchar(20);"`


	// Рег. номер банка
	RegNumber			int `json:"reg_number" gorm:"type:int;"`

	// Дата регистрации
	RegistrationDate	time.Time `json:"registration_date"`

	// Действующий / Ликвидируется / ..
	Status	string `json:"status" gorm:"type:varchar(32);"`

	// нужно ли?
	PaymentAccounts	[]PaymentAccount `json:"payment_accounts"`

	UpdatedAt 		time.Time `json:"updated_at"`
}

func (Bank) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&Bank{}); err != nil {
		log.Fatal(err)
	}
	err := db.Exec("ALTER TABLE banks " +
		"ADD CONSTRAINT banks_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

	// 2.
	units := []Bank{
		{AccountId: 1, Name:"АО \"АЛЬФА-БАНК\"", 	Address: "107078, г Москва, ул Каланчевская, д 27", RCBIC: "044525593", SWIFT: "ALFARUMMXXX",INN: "7728168971",	KPP: "770801001",CorporateAccount: "30101810200000000593",RegNumber: 1326,RegistrationDate: time.Date(1991,1,3,1,0,0,0,time.UTC), Status: "Действующий"},
		{AccountId: 1, Name:"ПАО СБЕРБАНК", 		Address: "117312, г Москва, ул Вавилова, д 19", 	RCBIC: "044525225", SWIFT: "SABRRUMM",	INN: "7707083893",	KPP: "773601001",CorporateAccount: "30101810400000000225",RegNumber: 1481,RegistrationDate: time.Date(1991,6,20,1,0,0,0,time.UTC), Status: "Действующий"},
	}

	for i, _ := range units {
		_, err := units[i].create()
		if err != nil {
			fmt.Println("Cannot create UnitMeasurement: ", err)
		}
	}
}
func (bank *Bank) BeforeCreate(tx *gorm.DB) error {
	bank.Id = 0
	return nil
}
func (bank *Bank) AfterFind(tx *gorm.DB) (err error) {

	return nil
}
func (bank *Bank) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(bank)
	} else {
		_db = _db.Model(&Bank{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"PaymentAccount"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}
}

// ############# Entity interface #############
func (bank Bank) GetId() uint { return bank.Id }
func (bank *Bank) setId(id uint) { bank.Id = id }
func (bank *Bank) setPublicId(publicId uint) { }
func (bank Bank) GetAccountId() uint { return bank.AccountId }
func (bank *Bank) setAccountId(id uint) { bank.AccountId = id }
func (bank Bank) SystemEntity() bool { return bank.AccountId == 1 }
// ############# End Entity interface #############

// ######### CRUD Functions ############
func (bank Bank) create() (Entity, error)  {

	en := bank

	if err := db.Create(&en).Error; err != nil {
		return nil, err
	}

	err := en.GetPreloadDb(false, false, nil).First(&en, en.Id).Error
	if err != nil {
		return nil, err
	}

	var newItem Entity = &en

	return newItem, nil
}
func (Bank) get(id uint, preloads []string) (Entity, error) {

	var bank Bank

	err := bank.GetPreloadDb(false,false,preloads).First(&bank, id).Error
	if err != nil {
		return nil, err
	}
	return &bank, nil
}
func (bank *Bank) load(preloads []string) error {

	err := bank.GetPreloadDb(false,false,preloads).First(bank, bank.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (bank *Bank) loadByPublicId(preloads []string) error {
	
	return utils.Error{Message: "Модель не может быть загружена через публичный id"}
}
func (Bank) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return Bank{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (Bank) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	banks := make([]Bank,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&Bank{}).GetPreloadDb(false,false,preloads).Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id IN (?)", []uint{1, accountId}).
			Where(filter).Find(&banks, "name ILIKE ?", search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Bank{}).
			Where("account_id IN (?) AND name ILIKE ?", []uint{1, accountId}, search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		
		err := (&Bank{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id IN (?)", []uint{1, accountId}).
			Where(filter).Find(&banks).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Bank{}).Where("account_id IN (?)", []uint{1, accountId}).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(banks))
	for i := range banks {
		entities[i] = &banks[i]
	}

	return entities, total, nil
}
func (bank *Bank) update(input map[string]interface{}, preloads []string) error {

	delete(input,"payment_accounts")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id"}); err != nil {
		return err
	}
	input = utils.FixInputDataTimeVars(input,[]string{"registration_date"})

	if err := bank.GetPreloadDb(false,false,nil).Where(" id = ?", bank.Id).
		Omit("id", "account_id","public_id").Updates(input).Error; err != nil {
		return err
	}

	err := bank.GetPreloadDb(false,false,preloads).First(bank, bank.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (bank *Bank) delete () error {
	return bank.GetPreloadDb(true,false,nil).Where("id = ?", bank.Id).Delete(bank).Error
}
// ######### END CRUD Functions ############
