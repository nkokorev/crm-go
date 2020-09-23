package models

import (
	"database/sql"
	"fmt"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)

// Тег товара, статьи, может чего-то еще
type ProductTag struct {
	Id        			uint 	`json:"id" gorm:"primaryKey"`
	PublicId			uint   	`json:"public_id" gorm:"type:int;index;not null;"`
	AccountId 			uint 	`json:"-" gorm:"type:int;index;not null;"` // потребуется, если productGroupId == null

	Label	 			*string `json:"label" gorm:"type:varchar(255);"` 	// [пуэр,зеленый, красный, белый, улун], [лето,зима,осень,весна]
	Code	 			*string `json:"code" gorm:"type:varchar(255);"` 	// код, значение,..
	Color 				*string `json:"color" gorm:"type:varchar(32);"` 	// цвет чего-либо

	ProductTagGroupId	*uint 				`json:"product_tag_group_id" gorm:"type:int;"`
	ProductTagGroup		*ProductTagGroup	`json:"product_tag_group"`

	// число тегов *hidden*
	ProductCount 		int64 	`json:"_product_count" gorm:"-"`

	Products 			[]Product	`json:"products" gorm:"many2many:product_tag_products;"`
}

func (ProductTag) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&ProductTag{}); err != nil { log.Fatal(err) }
	err := db.Exec("ALTER TABLE product_tags \n    ADD CONSTRAINT product_tags_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    DROP CONSTRAINT IF EXISTS fk_product_tag_groups_tags,\n    DROP CONSTRAINT IF EXISTS fk_product_tags_product_tag_group;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

	err = db.SetupJoinTable(&ProductTag{}, "Products", &ProductTagProduct{})
	if err != nil {
		log.Fatal(err)
	}
}
func (productTag *ProductTag) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(productTag)
	} else {
		_db = _db.Model(&ProductTag{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"Products","ProductTagGroup"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}
}

// ############# Entity interface #############
func (productTag ProductTag) GetId() uint { return productTag.Id }
func (productTag *ProductTag) setId(id uint) { productTag.Id = id }
func (productTag *ProductTag) setPublicId(publicId uint) { productTag.PublicId = publicId }
func (productTag ProductTag) GetAccountId() uint { return productTag.AccountId }
func (productTag *ProductTag) setAccountId(id uint) { productTag.AccountId = id }
func (ProductTag) SystemEntity() bool { return false }
// ############# End Entity interface #############

func (productTag *ProductTag) BeforeCreate(tx *gorm.DB) error {
	productTag.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&ProductTag{}).Where("account_id = ?",  productTag.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	productTag.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (productTag *ProductTag) AfterFind(tx *gorm.DB) (err error) {
	productTag.ProductCount =  db.Model(productTag).Association("Products").Count()
	return nil
}
func (productTag *ProductTag) AfterCreate(tx *gorm.DB) error {
	return nil
}
func (productTag *ProductTag) AfterUpdate(tx *gorm.DB) error {
	return nil
}
func (productTag *ProductTag) AfterDelete(tx *gorm.DB) error {
	return nil
}

// ######### CRUD Functions ############
func (productTag ProductTag) create() (Entity, error)  {

	_item := productTag
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}
func (ProductTag) get(id uint, preloads []string) (Entity, error) {

	var productTag ProductTag

	err := productTag.GetPreloadDb(false,false,preloads).First(&productTag, id).Error
	if err != nil {
		return nil, err
	}
	return &productTag, nil
}
func (productTag *ProductTag) load(preloads []string) error {

	err := productTag.GetPreloadDb(false,false,preloads).First(productTag, productTag.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (productTag *ProductTag) loadByPublicId(preloads []string) error {

	if productTag.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить ProductTag - не указан  Id"}
	}
	if err := productTag.GetPreloadDb(false,false, preloads).First(productTag, "account_id = ? AND public_id = ?", productTag.AccountId, productTag.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (ProductTag) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return ProductTag{}.getPaginationList(accountId, 0, 25, sortBy, "",nil, preload)
}
func (ProductTag) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	productTags := make([]ProductTag,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&ProductTag{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order("product_tags." + sortBy).
			Joins("left join product_tag_groups on product_tag_groups.id = product_tags.product_tag_group_id").
			Select("product_tag_groups.*,product_tags.*").
			Where("product_tags.account_id = ? AND product_tag_groups.account_id = ? ", accountId, accountId).
			Find(&productTags, "product_tags.label ILIKE ? OR product_tags.code ILIKE ? OR product_tags.color ILIKE ? OR product_tag_groups.code ILIKE ?OR product_tag_groups.label ILIKE ?", search,search,search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			fmt.Println(err)
			return nil, 0, err
		}

		// Определяем total
		err = (&ProductTag{}).GetPreloadDb(false, false, nil).
			Joins("left join product_tag_groups on product_tag_groups.id = product_tags.product_tag_group_id").
			Select("product_tag_groups.*,product_tags.*").
			Where("product_tags.account_id = ? AND product_tag_groups.account_id = ? ", accountId, accountId).
			Where("product_tags.label ILIKE ? OR product_tags.code ILIKE ? OR product_tags.color ILIKE ? OR product_tag_groups.code ILIKE ?OR product_tag_groups.label ILIKE ?", search,search,search,search,search).
			Count(&total).Error
		if err != nil {
			fmt.Println(err)
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
		
	} else {

		err := (&ProductTag{}).GetPreloadDb(false,false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&productTags).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&ProductTag{}).GetPreloadDb(false,false, nil).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(productTags))
	for i := range productTags {
		entities[i] = &productTags[i]
	}

	return entities, total, nil
}
func (productTag *ProductTag) update(input map[string]interface{}, preloads []string) error {

	delete(input,"products")
	delete(input,"product_tag_group")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id","product_tag_group_id"}); err != nil {
		return err
	}
	// input = utils.FixInputDataTimeVars(input,[]string{"expired_at"})

	if err := productTag.GetPreloadDb(false,false,nil).Where(" id = ?", productTag.Id).
		Omit("id", "account_id","public_id").Updates(input).Error; err != nil {
		return err
	}

	err := productTag.GetPreloadDb(false,false,preloads).First(productTag, productTag.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (productTag *ProductTag) delete () error {
	return productTag.GetPreloadDb(true,false,nil).Where("id = ?", productTag.Id).Delete(productTag).Error
}
// ######### END CRUD Functions ############


////////////////
func (productTag ProductTag) GetProductPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	products := make([]Product,0)
	var total int64

	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&Product{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order("products."+sortBy).
			Joins("left join product_tag_products on product_tag_products.product_id = products.id").
			Joins("left join product_tags on product_tags.id = product_tag_products.product_tag_id").
			Select("product_tag_products.*,product_tags.*,products.*").
			Where("product_tags.id = ? AND products.account_id = ? AND product_tags.account_id = ?", productTag.Id, accountId, accountId).
			Find(&products, "products.label ILIKE ? OR products.short_label ILIKE ? OR products.article ILIKE ? OR products.brand ILIKE ? OR products.model ILIKE ? OR products.description ILIKE ?", search,search,search,search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			fmt.Println(err)
			return nil, 0, err
		}

		// Определяем total
		err = (&Product{}).GetPreloadDb(false, false, nil).
			Joins("left join product_tag_products on product_tag_products.product_id = products.id").
			Joins("left join product_tags on product_tags.id = product_tag_products.product_tag_id").
			Select("product_tag_products.*,product_tags.*,products.*").
			Where("product_tags.id = ? AND products.account_id = ? AND product_tags.account_id = ?", productTag.Id, accountId, accountId).
			Where("products.account_id = ? AND products.label ILIKE ? OR products.short_label ILIKE ? OR products.article ILIKE ? OR products.brand ILIKE ? OR products.model ILIKE ? OR products.description ILIKE ?", accountId, search,search,search,search,search,search).
			Count(&total).Error
		if err != nil {
			fmt.Println(err)
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {


		err := (&Product{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order("products."+sortBy).
			Joins("left join product_tag_products on product_tag_products.product_id = products.id").
			Joins("left join product_tags on product_tags.id = product_tag_products.product_tag_id").
			Select("product_tag_products.*,product_tags.*,products.*").
			Where("product_tags.id = ? AND products.account_id = ? AND product_tags.account_id = ?", productTag.Id, accountId, accountId).
			Find(&products).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			fmt.Println(err)
			return nil, 0, err
		}

		// Определяем total
		err = (&Product{}).GetPreloadDb(false, false, nil).
			Joins("left join product_tag_products on product_tag_products.product_id = products.id").
			Joins("left join product_tags on product_tags.id = product_tag_products.product_tag_id").
			Select("product_tag_products.*,product_tags.*,products.*").
			Where("product_tags.id = ? AND products.account_id = ? AND product_tags.account_id = ?", productTag.Id, accountId, accountId).
			Count(&total).Error
		if err != nil {
			fmt.Println(err)
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	entities := make([]Entity, len(products))
	for i := range products {
		entities[i] = &products[i]
	}

	return entities, total, nil
}

func (productTag *ProductTag) AppendProduct(product *Product, strict... bool) error {

	// 1. Загружаем продукт еще раз
	if err := product.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: нельзя добавить продукт, он не найден"}
	}

	// 2. Проверяем есть ли уже в этой категории этот продукт
	if productTag.ExistProductById(product.Id) {
		if len(strict) > 0 && strict[0] {
			return utils.Error{Message: "Продукт уже числиться за тегом"}
		} else {
			return nil
		}
	}

	if err := db.Create(
		&ProductTagProduct{
			ProductId: product.Id, ProductTagId: productTag.Id}).Error; err != nil {
		return err
	}

	return nil
}
func (productTag *ProductTag) RemoveProduct(product *Product) error {
	// 1. Загружаем продукт еще раз
	if err := product.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: нельзя удалить продукт, он не найден"}
	}

	if product.Id < 1 || productTag.Id < 1 {
		return utils.Error{Message: "Техническая ошибка: account id || product id || product category id == nil"}
	}

	if err := db.Where("product_tag_id = ? AND product_id = ?", productTag.Id, product.Id).Delete(
		&ProductTagProduct{}).Error; err != nil {
		return err
	}

	return nil
}
func (productTag *ProductTag) SyncProductByIds(products []Product) error {

	// Сначала очищаем список
	if err := db.Model(productTag).Association("Products").Clear(); err != nil {
		return err
	}

	for _,_product := range products {

		if err := productTag.AppendProduct(&Product{Id: _product.Id, AccountId: _product.AccountId}); err != nil {
			return err
		}

	}
	
	return nil
}
func (productTag *ProductTag) ExistProductById(productId uint) bool {

	var el ProductTagProduct

	err := db.Model(&ProductTagProduct{}).Where("product_tag_id = ? AND product_id = ?",productTag.Id, productId).First(&el).Error
	if err != nil {
		return false
	}

	return true
}