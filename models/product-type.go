package models

import (
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)

// Вид номенклатуры - ассортиментные группы продаваемых товаров.
type ProductType struct {
	Id     		uint	`json:"id" gorm:"primaryKey"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	// ulun, corner, ... (может нужно кому)
	Code 		*string `json:"code" gorm:"type:varchar(255);"`

	// Улунский, угло-зачистной, шерстянной и т.д.
	Name 		*string `json:"name" gorm:"type:varchar(255);"` 		// menu label - Чай, кофе, ..
	
	Products 	[]Product 	`json:"products" gorm:"ForeignKey:ProductTypeId;References:id;"`
}

// ############# Entity interface #############
func (productType ProductType) GetId() uint { return productType.Id }
func (productType *ProductType) setId(id uint) { productType.Id = id }
func (productType *ProductType) setPublicId(publicId uint) { }
func (productType ProductType) GetAccountId() uint { return productType.AccountId }
func (productType *ProductType) setAccountId(id uint) { productType.AccountId = id }
func (ProductType) SystemEntity() bool { return false }
// ############# End Entity interface #############
func (ProductType) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&ProductType{}); err != nil {
		log.Fatal(err)
	}
	// db.Model(&ProductType{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&ProductType{}).AddForeignKey("email_template_id", "email_templates(id)", "SET NULL", "CASCADE")
	// db.Model(&ProductType{}).AddForeignKey("email_box_id", "email_boxes(id)", "SET NULL", "CASCADE")
	// db.Model(&ProductType{}).AddForeignKey("users_segment_id", "users_segments(id)", "SET NULL", "CASCADE")
	err := db.Exec("ALTER TABLE product_types " +
		"ADD CONSTRAINT product_types_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (productType *ProductType) BeforeCreate(tx *gorm.DB) error {
	productType.Id = 0

	return nil
}
func (productType *ProductType) AfterFind(tx *gorm.DB) (err error) {

	return nil
}
// ######### CRUD Functions ############
func (productType ProductType) create() (Entity, error)  {

	en := productType

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
func (ProductType) get(id uint, preloads []string) (Entity, error) {

	var productType ProductType

	err := productType.GetPreloadDb(false,false,preloads).First(&productType, id).Error
	if err != nil {
		return nil, err
	}
	return &productType, nil
}
func (productType *ProductType) load(preloads []string) error {

	err := productType.GetPreloadDb(false,false,preloads).First(productType, productType.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (productType *ProductType) loadByPublicId(preloads []string) error {
	
	return utils.Error{Message: "Невозможно загрузить тип продукта через публичный ID"}
}
func (ProductType) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return ProductType{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (ProductType) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	productGroups := make([]ProductType,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&ProductType{}).GetPreloadDb(false,false,preloads).Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&productGroups, "label ILIKE ? OR code ILIKE ? OR route_name ILIKE ? OR icon_name ILIKE ? OR meta_title ILIKE ? OR meta_description ILIKE ? OR short_description ILIKE ? OR description ILIKE ?", search,search,search,search,search,search,search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&ProductType{}).
			Where("label ILIKE ? OR code ILIKE ? OR route_name ILIKE ? OR icon_name ILIKE ? OR meta_title ILIKE ? OR meta_description ILIKE ? OR short_description ILIKE ? OR description ILIKE ?", search,search,search,search,search,search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		
		err := (&ProductType{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&productGroups).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&ProductType{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(productGroups))
	for i := range productGroups {
		entities[i] = &productGroups[i]
	}

	return entities, total, nil
}
func (productType *ProductType) update(input map[string]interface{}, preloads []string) error {

	delete(input,"image")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"parent_id","priority","web_site_id"}); err != nil {
		return err
	}
	input = utils.FixInputDataTimeVars(input,[]string{"expired_at"})

	if err := productType.GetPreloadDb(false,false,nil).Where(" id = ?", productType.Id).
		Omit("id", "account_id","created_at","public_id").Updates(input).Error; err != nil {
		return err
	}

	err := productType.GetPreloadDb(false,false,preloads).First(productType, productType.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (productType *ProductType) delete () error {
	return productType.GetPreloadDb(true,false,nil).Where("id = ?", productType.Id).Delete(productType).Error
}
// ######### END CRUD Functions ############

func (productType *ProductType) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(productType)
	} else {
		_db = _db.Model(&ProductType{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"Products"})

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

