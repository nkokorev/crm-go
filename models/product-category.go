package models

import (
	"database/sql"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)

// Вид номенклатуры - ассортиментные группы продаваемых товаров.
// Product <> ProductCard <> ProductCategory <> WebPage .
// Смысл в том, что одна страница может объединять несколько Категорий: "Новинки"; "Популярное".
type ProductCategory struct {
	Id     		uint	`json:"id" gorm:"primaryKey"`
	PublicId	uint	`json:"public_id" gorm:"type:int;index;not null;"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`
	ParentId 	*uint	`json:"parent_id"`
	Children	[]ProductCategory `json:"children" gorm:"foreignkey:ParentId"`

	// Наименование категории: Шу Пуэр, Красный чай, Пуэр, ...
	Label 		*string `json:"label" gorm:"type:varchar(255);"`
	
	// Код категории для выборки (возможно), unique ( в рамках account)
	Code 		*string `json:"code" gorm:"type:varchar(255);"`

	// Приоритет отображения категории на странице
	// todo: доработать формат отображения... смешанный, по порядку и т.д.
	Priority	int		`json:"priority" gorm:"type:int;default:10;"`

	// Карточки товаров
	ProductCards 	[]ProductCard 	`json:"product_cards" gorm:"many2many:product_category_product_cards;"`

	ProductCardsCount 	uint 	`json:"_product_cards_count" gorm:"-"`
	WebPagesCount 		uint 	`json:"_web_pages_count" gorm:"-"`
	ProductsCount 		uint 	`json:"_products_count" gorm:"-"`

	// Страницы, на которых выводятся карточки товаров этой товарной группы
	WebPages 		[]WebPage 	`json:"web_pages" gorm:"many2many:web_page_product_categories;"`

	// Товары, которые входят в эту категорию
	Products 		[]Product 	`json:"products" gorm:"many2many:product_category_products;"`
}
func (ProductCategory) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&ProductCategory{}); err != nil {
		log.Fatal(err)
	}
	err := db.Exec("ALTER TABLE product_categories " +
		"ADD CONSTRAINT product_categories_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

	err = db.SetupJoinTable(&ProductCategory{}, "ProductCards", &ProductCategoryProductCard{})
	if err != nil {
		log.Fatal(err)
	}
	err = db.SetupJoinTable(&ProductCategory{}, "WebPages", &WebPageProductCategories{})
	if err != nil {
		log.Fatal(err)
	}
	err = db.SetupJoinTable(&ProductCategory{}, "Products", &ProductCategoryProduct{})
	if err != nil {
		log.Fatal(err)
	}
}
func (productCategory *ProductCategory) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(productCategory)
	} else {
		_db = _db.Model(&ProductCategory{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"Products","ProductCards","WebPages","Children","Children12"})

		for _,v := range allowed {
			if v == "Children12" {
				_db.Preload("Children.Children.Children.Children.Children.Children.Children.Children.Children.Children.Children.Children")
					
			} else {
				_db.Preload(v)
			}

		}
		return _db
	}
}

// ############# Entity interface #############
func (productCategory ProductCategory) GetId() uint { return productCategory.Id }
func (productCategory *ProductCategory) setId(id uint) { productCategory.Id = id }
func (productCategory *ProductCategory) setPublicId(publicId uint) { productCategory.PublicId = publicId }
func (productCategory ProductCategory) GetAccountId() uint { return productCategory.AccountId }
func (productCategory *ProductCategory) setAccountId(id uint) { productCategory.AccountId = id }
func (ProductCategory) SystemEntity() bool { return false }
// ############# End Entity interface #############

func (productCategory *ProductCategory) BeforeCreate(tx *gorm.DB) error {
	productCategory.Id = 0
	
	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&ProductCategory{}).Where("account_id = ?",  productCategory.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	productCategory.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (productCategory *ProductCategory) AfterFind(tx *gorm.DB) (err error) {

	// tx.Joins("ProductCategory").Preload("Children")
	// Считаем число товаров

	// 1. Собираем все ID дочерних категорий
	/*var arr []uint

	if productCategory.Children != nil {

		children := productCategory.Children
		for i := range children {
			child := children[i]
			arr = append(arr, child.Id)
		}
	}*/

	// fmt.Println("Ищем еще товары")
	

	/*stat := struct {
		ProductCardsCount uint
		WebPagesCount uint
		ProductsCount uint
	}{0,0,0}
	if err = db.Raw("SELECT\n    COUNT(*) AS units,\n    sum(payment_amount * volume_order) AS amount_order,\n    sum(payment_amount * volume_fact) AS amount_fact\nFROM shipment_items\nWHERE account_id = ? AND shipment_id = ?;",
		shipment.AccountId, shipment.Id).
		Scan(&stat).Error; err != nil {
		return err
	}

	productCategory.ProductCardsCount 	= stat.ProductCardsCount
	productCategory.WebPagesCount 		= stat.WebPagesCount
	productCategory.ProductsCount 		= stat.ProductsCount*/

	return nil
}
// ######### CRUD Functions ############
func (productCategory ProductCategory) create() (Entity, error)  {

	en := productCategory

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
func (ProductCategory) get(id uint, preloads []string) (Entity, error) {

	var productCategory ProductCategory

	err := productCategory.GetPreloadDb(false,false,preloads).First(&productCategory, id).Error
	if err != nil {
		return nil, err
	}
	return &productCategory, nil
}
func (productCategory *ProductCategory) load(preloads []string) error {

	err := productCategory.GetPreloadDb(false,false,preloads).First(productCategory, productCategory.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (productCategory *ProductCategory) loadByPublicId(preloads []string) error {
	
	if productCategory.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить ProductCategory - не указан  Id"}
	}
	if err := productCategory.GetPreloadDb(false,false, preloads).First(productCategory, "account_id = ? AND public_id = ?", productCategory.AccountId, productCategory.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (ProductCategory) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return ProductCategory{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (ProductCategory) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	productGroups := make([]ProductCategory,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&ProductCategory{}).GetPreloadDb(false,false,preloads).Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&productGroups, "label ILIKE ? OR code ILIKE ? OR route_name ILIKE ? OR icon_name ILIKE ? OR meta_title ILIKE ? OR meta_description ILIKE ? OR short_description ILIKE ? OR description ILIKE ?", search,search,search,search,search,search,search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&ProductCategory{}).
			Where("label ILIKE ? OR code ILIKE ? OR route_name ILIKE ? OR icon_name ILIKE ? OR meta_title ILIKE ? OR meta_description ILIKE ? OR short_description ILIKE ? OR description ILIKE ?", search,search,search,search,search,search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		
		err := (&ProductCategory{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&productGroups).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&ProductCategory{}).Where("account_id = ?", accountId).Count(&total).Error
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
func (productCategory *ProductCategory) update(input map[string]interface{}, preloads []string) error {

	delete(input,"image")
	delete(input,"web_pages")
	delete(input,"products")
	delete(input,"product_cards")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"parent_id","priority","web_site_id"}); err != nil {
		return err
	}
	input = utils.FixInputDataTimeVars(input,[]string{"expired_at"})

	if err := productCategory.GetPreloadDb(false,false,nil).Where(" id = ?", productCategory.Id).
		Omit("id", "account_id","created_at","public_id").Updates(input).Error; err != nil {
		return err
	}

	err := productCategory.GetPreloadDb(false,false,preloads).First(productCategory, productCategory.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (productCategory *ProductCategory) delete () error {
	return productCategory.GetPreloadDb(true,false,nil).Where("id = ?", productCategory.Id).Delete(productCategory).Error
}
// ######### END CRUD Functions ############


func (productCategory ProductCategory) CreateChild(wp ProductCategory) (Entity, error){
	wp.ParentId = utils.UINTp(productCategory.Id)
	wp.AccountId = productCategory.AccountId
	_webPage, err := wp.create()
	if err != nil {return nil, err}

	return _webPage, nil
}
func (productCategory ProductCategory) AppendProductCard(productCard *ProductCard, optPriority... int) error {

	priority := 10
	if len(optPriority) > 0 {
		priority = optPriority[0]
	}

	if productCard.Id < 1 {
		return utils.Error{Message: "Не создана продуктовая карточка"}
	}


	if err := db.Model(&ProductCategoryProductCard{}).Create(
		&ProductCategoryProductCard{ProductCategoryId: productCategory.Id, ProductCardId: productCard.Id, Priority: priority}).Error; err != nil {
		return err
	}

	account, err := GetAccount(productCard.AccountId)
	if err == nil && account != nil {
		event.AsyncFire(Event{}.ProductCardUpdated(account.Id, productCard.Id))
		event.AsyncFire(Event{}.ProductCategoryUpdated(account.Id, productCard.Id))
	}

	return nil
}
func (productCategory ProductCategory) RemoveProductCard(productCard ProductCard) error {

	// Загружаем еще раз
	if err := productCard.load(nil); err != nil {
		return err
	}

	//

	if productCategory.AccountId < 1 || productCard.Id < 1  || productCategory.Id < 1 {
		return utils.Error{Message: "Техническая ошибка: account id || product card id || product category id == nil"}
	}

	if err := db.Where("account_id = ? AND product_card_id = ? AND product_category_id = ?", productCategory.AccountId, productCard.Id, productCategory.Id).
		Delete(&ProductCategoryProductCard{}).Error; err != nil {
		return err
	}


	account, err := GetAccount(productCard.AccountId)
	if err == nil && account != nil {
		event.AsyncFire(Event{}.ProductCardUpdated(account.Id, productCard.Id))
		event.AsyncFire(Event{}.ProductCategoryUpdated(account.Id, productCard.Id))
	}

	return nil
}

func (productCategory *ProductCategory) AppendProduct(product *Product, strict... bool) error {

	// 1. Загружаем продукт еще раз
	if err := product.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: нельзя добавить продукт, он не найден"}
	}

	// 2. Проверяем есть ли уже в этой категории этот продукт
	if productCategory.ExistProduct(product.Id) {
		if len(strict) > 0 {
			return utils.Error{Message: "Продукт уже числиться в категории"}
		} else {
			return nil
		}
	}

	if err := db.Create(
		&ProductCategoryProduct{
			ProductId: product.Id, ProductCategoryId: productCategory.Id}).Error; err != nil {
		return err
	}

	return nil
}
func (productCategory *ProductCategory) RemoveProduct(product *Product) error {

	// 1. Загружаем продукт еще раз
	if err := product.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: нельзя удалить продукт, он не найден"}
	}

	if productCategory.AccountId < 1 || product.Id < 1 || productCategory.Id < 1 {
		return utils.Error{Message: "Техническая ошибка: account id || product id || product category id == nil"}
	}

	if err := db.Where("product_category_id = ? AND product_id = ?", productCategory.Id, product.Id).Delete(
		&ProductCategoryProduct{}).Error; err != nil {
		return err
	}


	return nil
}

func (productCategory *ProductCategory) ExistProduct(productId uint) bool {

	if productId < 1 {
		return false
	}

	pcp := ProductCategoryProduct{}
	result := db.Model(&ProductCategoryProduct{}).First(&pcp,"product_category_id = ? AND product_id = ?", productCategory.Id, productId)

	if result.Error != nil {
		return false
	}
	if result.RowsAffected > 0 {
		return true
	}


	return false
}