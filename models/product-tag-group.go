package models

import (
	"database/sql"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"log"
)

// Карточка "товара" в магазине в котором могут быть разные торговые предложения
type ProductTagGroup struct {
	Id        			uint 	`json:"id" gorm:"primaryKey"`
	PublicId			uint   	`json:"public_id" gorm:"type:int;index;not null;"`
	AccountId 			uint 	`json:"-" gorm:"type:int;index;not null;"` // потребуется, если productGroupId == null

	// [пуэр,зеленый, красный, белый, улун], [лето,зима,осень,весна], [рассыпной, упаковка]
	Label	 			*string `json:"label" gorm:"type:varchar(255);"`
	Code	 			*string `json:"code" gorm:"type:varchar(255);"`

	// Filter, по которому можно фильтровать данные
	FilterLabel	 			*string `json:"filter_label" gorm:"type:varchar(255);"`
	FilterCode	 			*string `json:"filter_code" gorm:"type:varchar(255);"`

	// todo: добавить выбор доступных фильтров

	Color 				*string `json:"color" gorm:"type:varchar(32);"`

	// Что-то про фильтры
	EnableSorting	 	bool 	`json:"enable_sorting" gorm:"type:bool;default:true"`
	EnableView			bool 	`json:"enable_view" gorm:"type:bool;default:true"`

	// краткое описание группы тегов ()
	Description 	*string 	`json:"description" gorm:"type:varchar(255);"`

	// число тегов *hidden*
	TagCount 		int64 	`json:"_tag_count" gorm:"-"` 

	Tags 			[]ProductTag	`json:"tags"`
}

func (ProductTagGroup) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&ProductTagGroup{}); err != nil { log.Fatal(err) }
	err := db.Exec("ALTER TABLE product_tag_groups ADD CONSTRAINT product_tag_groups_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (productTagGroup *ProductTagGroup) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(productTagGroup)
	} else {
		_db = _db.Model(&ProductTagGroup{})
	}

	if autoPreload {
		return db.Preload("Products").Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Select(Storage{}.SelectArrayWithoutDataURL())
		})
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"Images","Products"})

		for _,v := range allowed {
			if v == "Images" {
				_db.Preload("Images", func(db *gorm.DB) *gorm.DB {
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
func (productTagGroup ProductTagGroup) GetId() uint { return productTagGroup.Id }
func (productTagGroup *ProductTagGroup) setId(id uint) { productTagGroup.Id = id }
func (productTagGroup *ProductTagGroup) setPublicId(publicId uint) { productTagGroup.PublicId = publicId }
func (productTagGroup ProductTagGroup) GetAccountId() uint { return productTagGroup.AccountId }
func (productTagGroup *ProductTagGroup) setAccountId(id uint) { productTagGroup.AccountId = id }
func (ProductTagGroup) SystemEntity() bool { return false }
// ############# End Entity interface #############

func (productTagGroup *ProductTagGroup) BeforeCreate(tx *gorm.DB) error {
	productTagGroup.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&ProductTagGroup{}).Where("account_id = ?",  productTagGroup.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	productTagGroup.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (productTagGroup *ProductTagGroup) AfterFind(tx *gorm.DB) (err error) {
	productTagGroup.TagCount =  db.Model(productTagGroup).Association("Tags").Count()
	return nil
}
func (productTagGroup *ProductTagGroup) AfterCreate(tx *gorm.DB) error {
	return nil
}
func (productTagGroup *ProductTagGroup) AfterUpdate(tx *gorm.DB) error {
	return nil
}
func (productTagGroup *ProductTagGroup) AfterDelete(tx *gorm.DB) error {
	return nil
}
// ######### CRUD Functions ############
func (productTagGroup ProductTagGroup) create() (Entity, error)  {

	_item := productTagGroup
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}
func (ProductTagGroup) get(id uint, preloads []string) (Entity, error) {

	var productTagGroup ProductTagGroup

	err := productTagGroup.GetPreloadDb(false,false,preloads).First(&productTagGroup, id).Error
	if err != nil {
		return nil, err
	}
	return &productTagGroup, nil
}
func (productTagGroup *ProductTagGroup) load(preloads []string) error {

	err := productTagGroup.GetPreloadDb(false,false,preloads).First(productTagGroup, productTagGroup.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (productTagGroup *ProductTagGroup) loadByPublicId(preloads []string) error {

	if productTagGroup.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить ProductTagGroup - не указан  Id"}
	}
	if err := productTagGroup.GetPreloadDb(false,false, preloads).First(productTagGroup, "account_id = ? AND public_id = ?", productTagGroup.AccountId, productTagGroup.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (ProductTagGroup) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return ProductTagGroup{}.getPaginationList(accountId, 0, 25, sortBy, "",nil, preload)
}
func (ProductTagGroup) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	productCards := make([]ProductTagGroup,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&ProductTagGroup{}).GetPreloadDb(false,false, preloads).
			Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).Where(filter).
			Find(&productCards, "label ILIKE ? OR path ILIKE ? OR breadcrumb ILIKE ? OR meta_title ILIKE ? OR meta_description ILIKE ? OR short_description ILIKE ? OR description ILIKE ?", search,search,search,search,search,search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&ProductTagGroup{}).GetPreloadDb(false,false, nil).
			Where("label ILIKE ? OR path ILIKE ? OR breadcrumb ILIKE ? OR meta_title ILIKE ? OR meta_description ILIKE ? OR short_description ILIKE ? OR description ILIKE ?", search,search,search,search,search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&ProductTagGroup{}).GetPreloadDb(false,false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&productCards).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&ProductTagGroup{}).GetPreloadDb(false,false, nil).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(productCards))
	for i := range productCards {
		entities[i] = &productCards[i]
	}

	return entities, total, nil
}
func (productTagGroup *ProductTagGroup) update(input map[string]interface{}, preloads []string) error {

	delete(input,"images")
	delete(input,"products")
	delete(input,"product_categories")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"parent_id","web_site_id","web_page_id"}); err != nil {
		return err
	}
	input = utils.FixInputDataTimeVars(input,[]string{"expired_at"})

	if err := productTagGroup.GetPreloadDb(false,false,nil).Where(" id = ?", productTagGroup.Id).
		Omit("id", "account_id","created_at","public_id").Updates(input).Error; err != nil {
		return err
	}

	err := productTagGroup.GetPreloadDb(false,false,preloads).First(productTagGroup, productTagGroup.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (productTagGroup *ProductTagGroup) delete () error {
	return productTagGroup.GetPreloadDb(true,false,nil).Where("id = ?", productTagGroup.Id).Delete(productTagGroup).Error
}
// ######### END CRUD Functions ############


////////////////

func (productTagGroup *ProductTagGroup) AppendProductTag(productTag *ProductTag) error {

	// 1. Загружаем продукт еще раз
	if err := productTag.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: нельзя добавить tag, т.к. он не найден"}
	}

	productTag.ProductTagGroupId = &productTagGroup.Id

	// if err := db.Model(productTag).Update("product_tag_group_id", &productTagGroup.Id).Error; err != nil {
	if err := db.Save(productTag).Error; err != nil {
		return err
	}
	
	return nil
}
func (productTagGroup *ProductTagGroup) RemoveProductTag(productTag *ProductTag) error {

	// 1. Загружаем продукт еще раз
	if err := productTag.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: нельзя добавить tag, т.к. он не найден"}
	}

	productTag.ProductTagGroupId = nil

	if err := db.Save(productTag).Error; err != nil {
		return err
	}

	return nil
}

/*func (productTagGroup *ProductTagGroup) SyncProductTagByIds(productTags []ProductTag) error {

	// 1. Удалим все Ids...
	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Model(&ProductTagGroup{}).
		Where("account_id = ? AND product_tag_group_id = ?", productTagGroup.AccountId, productTagGroup.Id ).
		Update("product_tag_group_id", nil).Error; err != nil {
			return err
	}

	for _,_productTag := range productTags {

		if err := productTagGroup.AppendTag(&ProductTag{Id: _productTag.Id, AccountId: productTagGroup.AccountId}); err != nil {
			return err
		}
	}

	return nil
}*/
