package models

import (
	"database/sql"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"log"
)

// Карточка "товара" в магазине в котором могут быть разные торговые предложения
type ProductCard struct {
	Id        			uint 	`json:"id" gorm:"primaryKey"`
	PublicId			uint   	`json:"public_id" gorm:"type:int;index;not null;"`
	AccountId 			uint 	`json:"-" gorm:"type:int;index;not null;"` // потребуется, если productGroupId == null
	WebSiteId 			*uint 	`json:"web_site_id" gorm:"type:int;"` // магазин, к которому относится

	// Активна ли карточка товара
	EnableRetailSale 	bool 	`json:"enable_retail_sale" gorm:"type:bool;default:true"`
	EnableWholesaleSale bool 	`json:"enable_wholesale_sale" gorm:"type:bool;default:true"`

	Label	 			*string 	`json:"label" gorm:"type:varchar(255);"` 	// что выводить в список товаров
	SecondLabel 		*string 	`json:"second_label" gorm:"type:varchar(255);"`	// Второе название часто бывает
	Path 				*string 	`json:"path" gorm:"type:varchar(255);"` 	// идентификатор страницы (syao-chzhun )
	RouteName 			*string 	`json:"route_name" gorm:"type:varchar(50);"`    // {catalog.product} - может быть удобно в каких-то фреймворках
	Breadcrumb 			*string 	`json:"breadcrumb" gorm:"type:varchar(255);"`

	MetaTitle 			*string 	`json:"meta_title" gorm:"type:varchar(255);"`
	MetaKeywords 		*string 	`json:"meta_keywords" gorm:"type:varchar(255);"`
	MetaDescription 	*string 	`json:"meta_description" gorm:"type:varchar(255);"`

	// Full description нет т.к. в карточке описание берется от офера
	ShortDescription 	*string 	`json:"short_description" gorm:"type:varchar(255);"` // для превью карточки товара
	Description 		*string 	`json:"description" gorm:"type:text;"` // фулл описание товара

	// число товаров *hidden*
	ProductCount 		int64 	`json:"_product_count" gorm:"-"` 

	// Хелперы карточки: какой параметр выводить в качестве переключателя(ей) (цвета, шт, кг и т.д.)
	SwitchProducts	 	datatypes.JSON `json:"switch_products"` // {color, size} Параметры переключения среди предложений

	// Preview Images - небольшие пережатые изображения товара(ов)
	Images 				[]Storage	`json:"images" gorm:"polymorphic:Owner;"`

	ProductTags			[]ProductTag `json:"product_tags" gorm:"many2many:product_tag_product_cards;"`

	Products 			[]Product 		`json:"products" gorm:"many2many:product_card_products;ForeignKey:id;References:id;"`
}

func (ProductCard) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&ProductCard{}); err != nil { log.Fatal(err) }
	err := db.Exec("ALTER TABLE product_cards ADD CONSTRAINT product_cards_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

	err = db.SetupJoinTable(&ProductCard{}, "Products", &ProductCardProduct{})
	if err != nil {
		log.Fatal(err)
	}
}
func (productCard *ProductCard) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(productCard)
	} else {
		_db = _db.Model(&ProductCard{})
	}

	if autoPreload {                                                                                   
		return db.Preload("Products").Preload("Products.ProductCards").Preload("Products.PaymentSubject").Preload("Products.MeasurementUnit").
			Preload("ProductTags").
			Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Select(Storage{}.SelectArrayWithoutDataURL())
		})
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,
			[]string{"Images","Products","Products.ProductCards","Products.ProductTags","Products.PaymentSubject","Products.MeasurementUnit","ProductTags"})

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
func (productCard ProductCard) GetId() uint { return productCard.Id }
func (productCard *ProductCard) setId(id uint) { productCard.Id = id }
func (productCard *ProductCard) setPublicId(publicId uint) { productCard.PublicId = publicId }
func (productCard ProductCard) GetAccountId() uint { return productCard.AccountId }
func (productCard *ProductCard) setAccountId(id uint) { productCard.AccountId = id }
func (ProductCard) SystemEntity() bool { return false }
// ############# End Entity interface #############

func (productCard *ProductCard) BeforeCreate(tx *gorm.DB) error {
	productCard.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&ProductCard{}).Where("account_id = ?",  productCard.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	productCard.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (productCard *ProductCard) AfterFind(tx *gorm.DB) (err error) {
	productCard.ProductCount =  db.Model(productCard).Association("Products").Count()
	return nil
}
// ######### CRUD Functions ############
func (productCard ProductCard) create() (Entity, error)  {

	_item := productCard
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	AsyncFire(NewEvent("ProductCardCreated", map[string]interface{}{"account_id":_item.AccountId, "product_card_id":_item.Id}))

	var entity Entity = &_item

	return entity, nil
}
func (ProductCard) get(id uint, preloads []string) (Entity, error) {

	var productCard ProductCard

	err := productCard.GetPreloadDb(false,false,preloads).First(&productCard, id).Error
	if err != nil {
		return nil, err
	}
	return &productCard, nil
}
func (productCard *ProductCard) load(preloads []string) error {

	err := productCard.GetPreloadDb(false,false,preloads).First(productCard, productCard.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (productCard *ProductCard) loadByPublicId(preloads []string) error {

	if productCard.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить ProductCard - не указан  Id"}
	}
	if err := productCard.GetPreloadDb(false,false, preloads).First(productCard, "account_id = ? AND public_id = ?", productCard.AccountId, productCard.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (ProductCard) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return ProductCard{}.getPaginationList(accountId, 0, 25, sortBy, "",nil, preload)
}
func (ProductCard) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	productCards := make([]ProductCard,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&ProductCard{}).GetPreloadDb(false,false, preloads).
			Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).Where(filter).
			Find(&productCards, "label ILIKE ? OR path ILIKE ? OR breadcrumb ILIKE ? OR meta_title ILIKE ? OR meta_description ILIKE ? OR short_description ILIKE ? OR description ILIKE ?", search,search,search,search,search,search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&ProductCard{}).GetPreloadDb(false,false, nil).
			Where("label ILIKE ? OR path ILIKE ? OR breadcrumb ILIKE ? OR meta_title ILIKE ? OR meta_description ILIKE ? OR short_description ILIKE ? OR description ILIKE ?", search,search,search,search,search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&ProductCard{}).GetPreloadDb(false,false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&productCards).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&ProductCard{}).GetPreloadDb(false,false, nil).Where("account_id = ?", accountId).Count(&total).Error
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
func (productCard *ProductCard) update(input map[string]interface{}, preloads []string) error {

	delete(input,"images")
	delete(input,"image")
	delete(input,"products")
	delete(input,"product_categories")
	delete(input,"product_tags")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"parent_id","web_site_id","web_page_id"}); err != nil {
		return err
	}
	input = utils.FixInputDataTimeVars(input,[]string{"expired_at"})

	if err := productCard.GetPreloadDb(false,false,nil).Where(" id = ?", productCard.Id).
		Omit("id", "account_id","created_at","public_id").Updates(input).Error; err != nil {
		return err
	}

	AsyncFire(NewEvent("ProductCardUpdated", map[string]interface{}{"account_id":productCard.AccountId, "product_card_id":productCard.Id}))

	err := productCard.GetPreloadDb(false,false,preloads).First(productCard, productCard.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (productCard *ProductCard) delete () error {
	if err := productCard.GetPreloadDb(true,false,nil).Where("id = ?", productCard.Id).Delete(productCard).Error; err != nil {
		return err
	}
	AsyncFire(NewEvent("ProductCardDeleted", map[string]interface{}{"account_id":productCard.AccountId, "product_card_id":productCard.Id}))

	return nil
}
// ######### END CRUD Functions ############


////////////////

func (productCard *ProductCard) AppendProduct(input *Product, optPriority... int) error {

	priority := 10
	if len(optPriority) > 0 {
		priority = optPriority[0]
	}
	var product *Product
	if input.Id < 1 {
		proPtr, err := input.create()
		if err != nil {
			return err
		}
		_p,ok := proPtr.(*Product)
		if !ok {
			return utils.Error{Message: "Ошибка преобразования *Product"}
		}
		product = _p
	} else {

		// Продукт уже есть
		if productCard.ExistProductById(input.Id) {
			return nil
		}
		product = input
	}
	if err := db.Model(&ProductCardProduct{}).Create(&ProductCardProduct{ProductId: product.Id, ProductCardId: productCard.Id, Priority: priority}).Error; err != nil {
		return err
	}

	account, err := GetAccount(productCard.AccountId)
	if err == nil && account != nil {
		// AsyncFire(NewEvent("ProductCardUpdated", map[string]interface{}{"account_id":account.Id, "product_card_id":productCard.Id}))
		AsyncFire(NewEvent("ProductCardAppendedProduct",
			map[string]interface{}{"account_id":product.AccountId, "product_id":product.Id, "product_card_id":productCard.Id}))
	}
	
	return nil
}
func (productCard *ProductCard) RemoveProduct(product *Product) error {

	if product.Id < 1 {
		return utils.Error{Message: "Необходимо указать верный id товара"}
	}

	if err := db.Model(productCard).Association("Products").Delete(product); err != nil {
		return err
	}

	account, err := GetAccount(productCard.AccountId)
	if err == nil && account != nil {
		AsyncFire(NewEvent("ProductCardRemovedProduct",
			map[string]interface{}{"account_id":product.AccountId, "product_id":product.Id, "product_card_id":productCard.Id}))
	}

	return nil
}

func (productCard *ProductCard) SyncProductByIds(products []Product) error {

	// очищаем список
	if err := db.Model(productCard).Association("Products").Clear(); err != nil {
		return err
	}

	for _,_product := range products {

		if err := productCard.AppendProduct(&Product{Id: _product.Id, AccountId: _product.AccountId}, 10); err != nil {
			return err
		}

	}
	
	return nil
}


// спорная функция
func (productCard *ProductCard) ExistProductById(productId uint) bool {

	var el ProductCardProduct

	err := db.Model(&ProductCardProduct{}).Where("product_card_id = ? AND product_id = ?",productCard.Id, productId).First(&el).Error
	if err != nil {
		return false
	}

	return true
}
func (productCard ProductCard) ManyToManyProductById(productId uint) (*ProductCardProduct, error) {
	
	var el ProductCardProduct

	err := db.Model(&ProductCardProduct{}).Where("product_card_id = ? AND product_id = ?",productCard.Id, productId).First(&el).Error
	if err != nil {
		return nil, utils.Error{Message: "Товар не найден в карточке товара"}
	}

	return &el, nil
}

func (productCard *ProductCard) AppendProductTag(productTag *ProductTag, strict... bool) error {

	// 1. Загружаем продукт еще раз
	if err := productTag.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: нельзя добавить tag, она не найдена"}
	}

	if productCard.Id < 1 {
		return utils.Error{Message: "Техническая ошибка: нельзя добавить tag, т.к. продукта не загружен"}
	}

	// 2. Проверяем есть ли уже в этой категории этот продукт
	if productCard.ExistProductTag(productTag.Id) {
		if len(strict) > 0 && strict[0] {
			return utils.Error{Message: "Tag уже числиться за товаром"}
		} else {
			return nil
		}
	}

	if err := db.Create(
		&ProductTagProductCard{
			ProductCardId: productCard.Id, ProductTagId: productTag.Id}).Error; err != nil {
		return err
	}

	return nil
}
func (productCard *ProductCard) RemoveProductTag(productTag *ProductTag) error {

	// 1. Загружаем продукт еще раз
	if err := productTag.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: нельзя удалить продукт, он не найден"}
	}

	if productCard.Id < 1 || productTag.Id < 1 {
		return utils.Error{Message: "Техническая ошибка: account id || product id || product tag id == nil"}
	}

	if err := db.Where("product_tag_id = ? AND product_card_id = ?", productTag.Id, productCard.Id).Delete(
		&ProductTagProductCard{}).Error; err != nil {
		return err
	}

	return nil
}
func (productCard *ProductCard) SyncProductCardTagsByIds(ProductTags []ProductTag) error {

	// 1. Загружаем продукт еще раз
	if productCard.Id < 1 {
		return utils.Error{Message: "Тег не найден"}
	}

	// очищаем список категорий
	if err := db.Model(productCard).Association("ProductTags").Clear(); err != nil {
		return err
	}

	for _,_productTag := range ProductTags {
		if err := productCard.AppendProductTag(&ProductTag{Id: _productTag.Id}, false); err != nil {
			return err
		}
	}

	AsyncFire(NewEvent("ProductCardSyncProductTags", map[string]interface{}{"account_id":productCard.AccountId, "product_card_id":productCard.Id}))

	return nil
}
func (productCard *ProductCard) ExistProductTag(tagId uint) bool {

	if tagId < 1 {
		return false
	}

	pcp := ProductTagProductCard{}
	if err := db.Model(&ProductTagProductCard{}).First(&pcp,"product_tag_id = ? AND product_card_id = ?", tagId, productCard.Id).Error; err != nil {
		return false
	}

	return true
}