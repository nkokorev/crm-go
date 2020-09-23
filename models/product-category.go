package models

import (
	"database/sql"
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

	// Наименование категории в ед. числе: Шу Пуэр, Красный чай, Пуэр, ...
	Label 		*string `json:"label" gorm:"type:varchar(128);"`

	// Наименование категории в множ. числе: Шу Пуэры, Красные чаи, Пуэры, ...
	LabelPlural *string `json:"label_plural" gorm:"type:varchar(128);"`
	
	// Код категории для выборки (возможно), unique ( в рамках account), для назначения цвета
	Code 		*string `json:"code" gorm:"type:varchar(128);"`

	// Приоритет отображения категории на странице
	// todo: доработать формат отображения... смешанный, по порядку и т.д.
	Priority	int		`json:"priority" gorm:"type:int;default:10;"`

	// Отображать ли категорию в свойствах товара
	ShowProperty	bool	`json:"show_property" gorm:"type:bool;default:false"`

	ProductCardsCount 	uint 	`json:"_product_cards_count" gorm:"-"`
	WebPagesCount 		uint 	`json:"_web_pages_count" gorm:"-"`
	ProductsCount 		uint 	`json:"_products_count" gorm:"-"`
	
	// (up) Страницы, на которых выводятся карточки товаров этой товарной группы
	WebPages 		[]WebPage 	`json:"web_pages" gorm:"many2many:web_page_product_categories;"`

	// (down) Карточки товаров
	ProductCards 	[]ProductCard 	`json:"product_cards" gorm:"many2many:product_category_product_cards;"`

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

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"Products","WebPages","Children","Children12"})

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

	productCategories := make([]ProductCategory,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		// fmt.Println("search: ", search)
		// fmt.Println("limit: ", limit)
		// fmt.Println("offset: ", offset)
		// fmt.Println("sortBy: ", sortBy)
		// fmt.Println("accountId: ", accountId)

		err := (&ProductCategory{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&productCategories, "label ILIKE ? OR label_plural ILIKE ? OR code ILIKE ?", search,search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&ProductCategory{}).GetPreloadDb(false,false,nil).
			Where("account_id = ? AND label ILIKE ? OR label_plural ILIKE ? OR code ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		
		err := (&ProductCategory{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&productCategories).Error
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
	entities := make([]Entity,len(productCategories))
	for i := range productCategories {
		entities[i] = &productCategories[i]
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
func (productCategory *ProductCategory) AppendProduct(product *Product, strict... bool) error {

	// 1. Загружаем продукт еще раз
	if err := product.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: нельзя добавить продукт, он не найден"}
	}

	// 2. Проверяем есть ли уже в этой категории этот продукт
	if productCategory.ExistProduct(product.Id) {
		if len(strict) > 0 && strict[0] {
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

	if product.Id < 1 || productCategory.Id < 1 {
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