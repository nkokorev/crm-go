package models

import (
	"database/sql"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

// Контрагенты и клиенты
type Company struct {
	Id     		uint	`json:"id" gorm:"primaryKey"`
	PublicId	uint	`json:"public_id" gorm:"type:int;index;"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;"`

	// Полное имя компании: Общество с ограниченной ответственностью RatusMedia
	Name 		*string `json:"name" gorm:"type:varchar(255);"`
	// Сокращенное имя компании: ООО RatusMedia
	ShortName 	*string `json:"short_name" gorm:"type:varchar(128);"`

	// Основной вид деятельности
	PrimaryActivity	*string `json:"primary_activity" gorm:"type:varchar(255);"`

	// Сотрудники компании
	Users 	[]User `json:"users" gorm:"many2many:company_users;"`
	
	// ИНН (Taxpayer Identification Number)
	INN 	*uint `json:"inn" gorm:"type:int"`

	// КПП (Tax Registration Reason Code)
	KPP 	*uint `json:"kpp" gorm:"type:int"`

	// ОГРН (Primary State Registration Number)
	OGRN 	*uint `json:"ogrn" gorm:"type:int"`

	// Свидетельство о регистрации (у ИП)
	RegistrationCertificate	*string `json:"registration_certificate" gorm:"type:varchar(255);"`

	// Статус организации, действующая ли
	Status	bool `json:"status" gorm:"type:bool;default:true"`

	// Налоговая инспекция
	TaxInspection	*string `json:"tax_inspection" gorm:"type:varchar(255);"`

	// Телефон и почта 
	Phone	 	*string `json:"phone" gorm:"type:varchar(50);"`
	Email	 	*string `json:"email" gorm:"type:varchar(50);"`

	// Юридический адрес
	LegalAddress 	*string `json:"legal_address" gorm:"type:varchar(255);"`

	// Фактический адрес
	ActualAddress 	*string `json:"actual_address" gorm:"type:varchar(255);"`

	// Если это поставщик чего-либо и ему нужно красивое описание
	Description	*string `json:"description" gorm:"type:text;"`

	// Расчетные счета
	PaymentAccounts	[]PaymentAccount `json:"payment_accounts"`

	CreatedAt 	time.Time 		`json:"created_at"`
	UpdatedAt 	time.Time 		`json:"updated_at"`
	DeletedAt 	gorm.DeletedAt 	`json:"-" sql:"index"`
}

func (Company) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&Company{}); err != nil {
		log.Fatal(err)
	}
	err := db.Exec("ALTER TABLE companies " +
		"ADD CONSTRAINT companies_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

	err = db.SetupJoinTable(&Company{}, "Users", &CompanyUser{})
	if err != nil {
		log.Fatal(err)
	}
}
func (company *Company) BeforeCreate(tx *gorm.DB) error {
	company.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&Company{}).Where("account_id = ?",  company.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	company.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (company *Company) AfterFind(tx *gorm.DB) (err error) {
	return nil
}
func (company *Company) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(company)
	} else {
		_db = _db.Model(&Company{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"PaymentAccounts","Users"})

		for _,v := range allowed {
			if v == "Image" {
				_db.Preload("Image", func(db *gorm.DB) *gorm.DB {
					return db.Select(Storage{}.SelectArrayWithoutDataURL())
				})
			} else {
				_db.Preload(v)
			}
		}
		return _db
	}
}

// ############# Entity interface #############
func (company Company) GetId() uint { return company.Id }
func (company *Company) setId(id uint) { company.Id = id }
func (company *Company) setPublicId(publicId uint) { company.PublicId = publicId }
func (company Company) GetAccountId() uint { return company.AccountId }
func (company *Company) setAccountId(id uint) { company.AccountId = id }
func (Company) SystemEntity() bool { return false }
// ############# End Entity interface #############

// ######### CRUD Functions ############
func (company Company) create() (Entity, error)  {

	en := company

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
func (Company) get(id uint, preloads []string) (Entity, error) {

	var company Company

	err := company.GetPreloadDb(false,false,preloads).First(&company, id).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}
func (company *Company) load(preloads []string) error {

	err := company.GetPreloadDb(false,false,preloads).First(company, company.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (company *Company) loadByPublicId(preloads []string) error {
	
	if company.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить Company - не указан  Id"}
	}
	if err := company.GetPreloadDb(false,false, preloads).First(company, "account_id = ? AND public_id = ?", company.AccountId, company.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (Company) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return Company{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (Company) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	companies := make([]Company,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&Company{}).GetPreloadDb(false,false,preloads).Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&companies, "name ILIKE ? OR short_name ILIKE ? OR phone ILIKE ? OR email ILIKE ? OR description ILIKE ?", search,search,search,search,search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Company{}).
			Where("name ILIKE ? OR short_name ILIKE ? OR phone ILIKE ? OR email ILIKE ? OR description ILIKE ?", search,search,search,search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		
		err := (&Company{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&companies).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Company{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(companies))
	for i := range companies {
		entities[i] = &companies[i]
	}

	return entities, total, nil
}
func (company *Company) update(input map[string]interface{}, preloads []string) error {

	delete(input,"users")
	delete(input,"payment_accounts")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id","inn","kpp","ogrn"}); err != nil {
		return err
	}

	if err := company.GetPreloadDb(false,false,nil).Where(" id = ?", company.Id).
		Omit("id", "account_id","created_at","public_id").Updates(input).Error; err != nil {
		return err
	}

	err := company.GetPreloadDb(false,false,preloads).First(company, company.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (company *Company) delete () error {
	return company.GetPreloadDb(true,false,nil).Where("id = ?", company.Id).Delete(company).Error
}
// ######### END CRUD Functions ############

// ######### COMPANY ACTIONS ############


// ######### END COMPANY Functions ############