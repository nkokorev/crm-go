package models

import (
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)

// M <> M
type CompanyUser struct {
	Id     		uint	`json:"id" gorm:"primaryKey"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	CompanyId 	uint	`json:"company_id" gorm:"type:int;index;"`
	UserId 		uint	`json:"user_id" gorm:"type:int;index;"`

	// Доля в компании (если есть)
	Share		*float64 `json:"share" gorm:"type:numeric;"`

	// ИНН (Taxpayer Identification Number)
	INN 		*uint `json:"inn" gorm:"type:int"`

	// Директор, Владелец, ...
	RoleId		uint	`json:"role_id" gorm:"type:int;"`
	Role		Role 	`json:"role"`

	Company 	Company `json:"company"`
	User 		User 	`json:"user"`
}

func (CompanyUser) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&CompanyUser{}); err != nil {
		log.Fatal(err)
	}
	err := db.Exec("ALTER TABLE company_users " +
		"ADD CONSTRAINT company_users_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT company_users_company_id_fkey FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT company_users_role_id_fkey FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT company_users_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (companyUser *CompanyUser) BeforeCreate(tx *gorm.DB) error {
	companyUser.Id = 0

	return nil
}
func (companyUser *CompanyUser) AfterFind(tx *gorm.DB) (err error) {

	return nil
}
func (companyUser *CompanyUser) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(companyUser)
	} else {
		_db = _db.Model(&CompanyUser{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"Company","User","Role"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}
}

// ############# Entity interface #############
func (companyUser CompanyUser) GetId() uint { return companyUser.Id }
func (companyUser *CompanyUser) setId(id uint) { companyUser.Id = id }
func (companyUser *CompanyUser) setPublicId(publicId uint) { }
func (companyUser CompanyUser) GetAccountId() uint { return companyUser.AccountId }
func (companyUser *CompanyUser) setAccountId(id uint) { companyUser.AccountId = id }
func (CompanyUser) SystemEntity() bool { return false }
// ############# End Entity interface #############

// ######### CRUD Functions ############
func (companyUser CompanyUser) create() (Entity, error)  {

	en := companyUser

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
func (CompanyUser) get(id uint, preloads []string) (Entity, error) {

	var companyUser CompanyUser

	err := companyUser.GetPreloadDb(false,false,preloads).First(&companyUser, id).Error
	if err != nil {
		return nil, err
	}
	return &companyUser, nil
}
func (companyUser *CompanyUser) load(preloads []string) error {

	err := companyUser.GetPreloadDb(false,false,preloads).First(companyUser, companyUser.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (companyUser *CompanyUser) loadByPublicId(preloads []string) error {
	return utils.Error{Message: "Нельзя загрузить CompanyUser, т.к. нет публичного ключа"}
}
func (CompanyUser) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return CompanyUser{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (CompanyUser) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	companyUsers := make([]CompanyUser,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&CompanyUser{}).GetPreloadDb(false,false,preloads).
			Joins("LEFT JOIN products ON products.id = shipment_products.product_id").
			Select("products.*, shipment_products.*").
			Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&companyUsers, "shipment_id ILIKE ?", search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&CompanyUser{}).
			Where("shipment_id ILIKE ? ", search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		
		err := (&CompanyUser{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&companyUsers).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&CompanyUser{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(companyUsers))
	for i := range companyUsers {
		entities[i] = &companyUsers[i]
	}

	return entities, total, nil
}
func (companyUser *CompanyUser) update(input map[string]interface{}, preloads []string) error {

	delete(input,"company")
	delete(input,"user")
	delete(input,"role")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"company_id","user_id","role_id"}); err != nil {
		return err
	}
	// input = utils.FixInputDataTimeVars(input,[]string{"expired_at"})

	if err := companyUser.GetPreloadDb(false,false,nil).Where(" id = ?", companyUser.Id).
		Omit("id", "account_id","created_at").Updates(input).Error; err != nil {
		return err
	}

	err := companyUser.GetPreloadDb(false,false,preloads).First(companyUser, companyUser.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (companyUser *CompanyUser) delete () error {
	return companyUser.GetPreloadDb(true,false,nil).Where("id = ?", companyUser.Id).Delete(companyUser).Error
}
// ######### END CRUD Functions ############
