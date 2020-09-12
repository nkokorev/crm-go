package models

import (
	"database/sql"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)

// Контрагенты и клиенты
type Manufacturer struct {
	Id     		uint	`json:"id" gorm:"primaryKey"`
	PublicId	uint	`json:"public_id" gorm:"type:int;index;"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;"`

	// Полное имя компании: Общество с ограниченной ответственностью RatusMedia
	Name 		*string `json:"name" gorm:"type:varchar(255);"`
	
	// Сокращенное имя компании: ООО RatusMedia
	ShortName 	*string `json:"short_name" gorm:"type:varchar(128);"`

	// Основной вид деятельности: 'производитель чайного сырья'
	PrimaryActivity	*string `json:"primary_activity" gorm:"type:varchar(255);"`

	// Адрес производителя
	Address 	*string `json:"address" gorm:"type:varchar(255);"`
	
	// Вебсайт производителя
	WebSite		*string `json:"web_site" gorm:"type:varchar(255);"`

	// Обновлять только через AppendImage, превью изображение. Может быть logo...
	Image 		*Storage	`json:"image" gorm:"polymorphic:Owner;"`

	// Если это поставщик чего-либо и ему нужно красивое описание
	ShortDescription	*string `json:"short_description" gorm:"type:varchar(255);"`
	Description			*string `json:"description" gorm:"type:text;"`

}

func (Manufacturer) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&Manufacturer{}); err != nil {
		log.Fatal(err)
	}
	err := db.Exec("ALTER TABLE manufacturers " +
		"ADD CONSTRAINT manufacturers_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (manufacturer *Manufacturer) BeforeCreate(tx *gorm.DB) error {
	manufacturer.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&Manufacturer{}).Where("account_id = ?",  manufacturer.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	manufacturer.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (manufacturer *Manufacturer) AfterFind(tx *gorm.DB) (err error) {
	return nil
}
func (manufacturer *Manufacturer) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(manufacturer)
	} else {
		_db = _db.Model(&Manufacturer{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"Image"})

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
func (manufacturer Manufacturer) GetId() uint { return manufacturer.Id }
func (manufacturer *Manufacturer) setId(id uint) { manufacturer.Id = id }
func (manufacturer *Manufacturer) setPublicId(publicId uint) { manufacturer.PublicId = publicId }
func (manufacturer Manufacturer) GetAccountId() uint { return manufacturer.AccountId }
func (manufacturer *Manufacturer) setAccountId(id uint) { manufacturer.AccountId = id }
func (Manufacturer) SystemEntity() bool { return false }
// ############# End Entity interface #############

// ######### CRUD Functions ############
func (manufacturer Manufacturer) create() (Entity, error)  {

	en := manufacturer

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
func (Manufacturer) get(id uint, preloads []string) (Entity, error) {

	var manufacturer Manufacturer

	err := manufacturer.GetPreloadDb(false,false,preloads).First(&manufacturer, id).Error
	if err != nil {
		return nil, err
	}
	return &manufacturer, nil
}
func (manufacturer *Manufacturer) load(preloads []string) error {

	err := manufacturer.GetPreloadDb(false,false,preloads).First(manufacturer, manufacturer.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (manufacturer *Manufacturer) loadByPublicId(preloads []string) error {
	
	if manufacturer.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить Manufacturer - не указан  Id"}
	}
	if err := manufacturer.GetPreloadDb(false,false, preloads).First(manufacturer, "account_id = ? AND public_id = ?", manufacturer.AccountId, manufacturer.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (Manufacturer) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return Manufacturer{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (Manufacturer) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	manufacturers := make([]Manufacturer,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&Manufacturer{}).GetPreloadDb(false,false,preloads).Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&manufacturers, "name ILIKE ? OR short_name ILIKE ? OR phone ILIKE ? OR email ILIKE ? OR description ILIKE ?", search,search,search,search,search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Manufacturer{}).
			Where("name ILIKE ? OR short_name ILIKE ? OR phone ILIKE ? OR email ILIKE ? OR description ILIKE ?", search,search,search,search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		
		err := (&Manufacturer{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&manufacturers).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Manufacturer{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(manufacturers))
	for i := range manufacturers {
		entities[i] = &manufacturers[i]
	}

	return entities, total, nil
}
func (manufacturer *Manufacturer) update(input map[string]interface{}, preloads []string) error {

	delete(input,"users")
	delete(input,"payment_accounts")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id","inn","kpp","ogrn"}); err != nil {
		return err
	}

	if err := manufacturer.GetPreloadDb(false,false,nil).Where(" id = ?", manufacturer.Id).
		Omit("id", "account_id","created_at","public_id").Updates(input).Error; err != nil {
		return err
	}

	err := manufacturer.GetPreloadDb(false,false,preloads).First(manufacturer, manufacturer.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (manufacturer *Manufacturer) delete () error {
	return manufacturer.GetPreloadDb(true,false,nil).Where("id = ?", manufacturer.Id).Delete(manufacturer).Error
}
// ######### END CRUD Functions ############

// ######### COMPANY ACTIONS ############


// ######### END COMPANY Functions ############