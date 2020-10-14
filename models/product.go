package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"log"
)

/*
Продукт - как единица товара или услуги. То что потом в чеке у пользователя.
Продукт может быть как шт., упак., так и сборным из других товаров.
Продукт может входить во множество web-карточек (витрин)
Список характеристик продукта не регламентируются, но удобно, когда он принадлежит какой-то группе с фикс. списком параметров.
*/

type Product struct {
	Id        		uint 	`json:"id" gorm:"primaryKey"`

	PublicId		uint   	`json:"public_id" gorm:"type:int;index;not null;"`
	AccountId 		uint 	`json:"-" gorm:"type:int;index;not null;"`

	// Доступен ли товар для продажи в розницу
	RetailSale 		bool 	`json:"retail_sale" gorm:"type:bool;default:false"`

	// Доступен ли товар для продажи оптом
	WholesaleSale	bool	`json:"wholesale_sale" gorm:"type:bool;default:false"`

	// Сборный ли товар? При нем warehouse_items >= 1. Применяется только к payment_subject = commodity, excise и т.д.
	IsKit			bool 		`json:"is_kit" gorm:"type:bool;default:false"`

	// При isSource = true, - из каких товаров и в каком количестве состоит
	Sources			[]*Product `json:"-" gorm:"many2many:product_sources;"` // ForeignKey:id;References:id;
	SourceItems		[]*ProductSource `json:"source_items"`

	// Этикетка товара
	Label 			*string 	`json:"label" gorm:"type:varchar(128);"`
	SecondLabel 	*string 	`json:"second_label" gorm:"type:varchar(128);"`	// Второе название часто бывает
	ShortLabel 		*string 	`json:"short_label" gorm:"type:varchar(128);"` // Краткое / radio.label

	// артикул товара
	Article 		*string 	`json:"article" gorm:"type:varchar(128);"`
	
	// торговая марка (Объект!)
	Trademark 		*string		`json:"trademark" gorm:"type:varchar(128);"`

	// Маркировка товара
	Brand 			*string		`json:"brand" gorm:"type:varchar(128);"`

	// Общая тема типа группы товаров, может повторяться для вывода в web-интерфейсе как "одного" товара
	Model 			*string		`json:"model" gorm:"type:varchar(255);"`

	// Base properties
	RetailPrice			*float64 `json:"retail_price" gorm:"type:numeric;"` 		// розничная цена
	
	WholesalePrice1 	*float64 `json:"wholesale_price_1" gorm:"type:numeric;column:wholesale_price_1;"` 	// оптовая цена
	WholesalePrice2 	*float64 `json:"wholesale_price_2" gorm:"type:numeric;column:wholesale_price_2;"` 	// оптовая цена
	WholesalePrice3 	*float64 `json:"wholesale_price_3" gorm:"type:numeric;column:wholesale_price_3;"` 	// оптовая цена
	
	RetailDiscount 		*float64 `json:"retail_discount" gorm:"type:numeric;"` 	// розничная фактическая скидка

	// Вид номенклатуры - ассортиментные группы продаваемых товаров. Привязываются к карточкам..

	// Товарная группа для назначения характеристик
	// PaymentGroupId	uint	`json:"payment_group_id" gorm:"type:int;"`
	// ProductGroup	ProductGroup `json:"product_group"`

	// Тип продукта: улунский, красный (чай), углозачистной станок, шлифовальный станок
	ProductTypeId		*uint		`json:"product_type_id" gorm:"type:int;"`
	ProductType			ProductType `json:"product_type"`

	ProductTags			[]ProductTag `json:"product_tags" gorm:"many2many:product_tag_products;"`

	// Список продуктов из которых составлен текущий. Это может быть как 1<>1, а может быть и нет (== составной товар)
	WarehouseItems		[]WarehouseItem `json:"warehouse_items"`
	Warehouses			[]Warehouse 	`json:"warehouses" gorm:"many2many:warehouse_item;"`

	// Ед. измерения товара: штуки, метры, литры, граммы и т.д.  !!!!
	MeasurementUnitId 	*uint	`json:"measurement_unit_id" gorm:"type:int;"` // тип измерения
	MeasurementUnit 	MeasurementUnit `json:"measurement_unit"`// Ед. измерения: штуки, коробки, комплекты, кг, гр, пог.м.

	// Целое или дробное количество товара
	IsInteger 			bool	`json:"is_integer" gorm:"type:bool;default:true"`

	// Основные атрибуты для расчета (Можно и в атрибуты)
	Length 	*float64 `json:"length" gorm:"type:numeric;"`
	Width 	*float64 `json:"width" gorm:"type:numeric;"`
	Height 	*float64 `json:"height" gorm:"type:numeric;"`
	Weight 	*float64 `json:"weight" gorm:"type:numeric;"`

	// Производитель (не поставщик)
	ManufacturerId	*uint		`json:"manufacturer_id" gorm:"type:int;"`
	Manufacturer	Manufacturer `json:"manufacturer"`

	// Дата изготовления (условная штука т.к. зависит от поставки), дата выпуска, дата производства
	ManufactureDate	*string 	`json:"manufacture_date" gorm:"type:varchar(255);"`

	// Условия хранения
	StorageRequirements	*string	`json:"storage_requirements" gorm:"type:varchar(255);"`

	// Срок годности, срок хранения (?)
	ShelfLife		*string 	`json:"shelf_life" gorm:"type:varchar(255);"`

	//  == признак предмета расчета - товар, услуга, работа, набор (комплект) = сборный товар
	// Признак предмета расчета (бухучет - № 54-ФЗ)
	PaymentSubjectId	*uint	`json:"payment_subject_id" gorm:"type:int;"`
	PaymentSubject 		PaymentSubject `json:"payment_subject"`
	
	// Ставка НДС или учет НДС (бухучет)
	VatCodeId	*uint	`json:"vat_code_id" gorm:"type:int;default:1;"`// товар или услуга ? [вид номенклатуры]
	VatCode		VatCode	`json:"vat_code"`

	ShortDescription 	*string 	`json:"short_description" gorm:"type:varchar(255);"` // pgsql: varchar - это зачем?)
	Description 		*string 	`json:"description" gorm:"type:text;"` // pgsql: text

	// Обновлять только через AppendImage
	Images 			[]Storage 	`json:"images" gorm:"polymorphic:Owner;"`
	
	Attributes 	datatypes.JSON `json:"attributes"`

	// todo: можно изменить и сделать свойства товара
	// ключ для расчета веса продукта
	// WeightKey 	string `json:"weight_key" gorm:"type:varchar(32);default:'grossWeight'"`

	// Весовой товар? = нужно ли считать вес для расчета доставки у данного продукта
	ConsiderWeight	bool	`json:"consider_weight" gorm:"type:bool;default:true"`

	// Reviews []Review // Product reviews (отзывы на товар - с рейтингом(?))
	// Questions []Question // вопросы по товару
	// Video []Video // видеообзоры по товару на ютубе

	// Список поставок товара, в которых он был
	Shipments 			[]Shipment 	`json:"shipments" gorm:"many2many:shipment_items"`
	ShipmentItems 		[]ShipmentItem 	`json:"shipment_items" gorm:"many2many:shipment_items"`

	Inventories 		[]Inventory	`json:"inventories" gorm:"many2many:inventory_items"`

	ProductCards 		[]ProductCard 	`json:"product_cards" gorm:"many2many:product_card_products;ForeignKey:id;References:id;"`
	ProductCategories 	[]ProductCategory 	`json:"product_categories" gorm:"many2many:product_category_products;"`
}

func (Product) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	if err := db.Migrator().CreateTable(&Product{}); err != nil {log.Fatal(err)}

	// db.Model(&Product{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE products ADD CONSTRAINT products_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
	db.Exec("create unique index uix_products_account_id_model ON products (account_id,model) WHERE (length(model) > 0);\ncreate unique index uix_products_account_id_article ON products (account_id,article) WHERE (length(article) > 0);\n")

	err = db.SetupJoinTable(&Product{}, "ProductCards", &ProductCardProduct{})
	if err != nil {
		log.Fatal(err)
	}

	err = db.SetupJoinTable(&Product{}, "Warehouses", &WarehouseItem{})
	if err != nil {
		log.Fatal(err)
	}

	err = db.SetupJoinTable(&Product{}, "Shipments", &ShipmentItem{})
	if err != nil {
		log.Fatal(err)
	}

	err = db.SetupJoinTable(&Product{}, "Inventories", &InventoryItem{})
	if err != nil {
		log.Fatal(err)
	}

	err = db.SetupJoinTable(&Product{}, "ProductCategories", &ProductCategoryProduct{})
	if err != nil {
		log.Fatal(err)
	}

	err = db.SetupJoinTable(&Product{}, "ProductTags", &ProductTagProduct{})
	if err != nil {
		log.Fatal(err)
	}
}

func (product *Product) BeforeCreate(tx *gorm.DB) error {
	product.Id = 0

	// 1. Рассчитываем PublicId (#id заказа) внутри аккаунта
	var lastIdx sql.NullInt64

	err := db.Model(&Product{}).Where("account_id = ?",  product.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	product.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (product *Product) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(product)
	} else {
		_db = _db.Model(&Product{})
	}

	if autoPreload {
		return db.Preload("ProductType","ProductCategories","PaymentSubject","VatCode","MeasurementUnit","Account","ProductCards",
			"Manufacturer","ProductTags","SourceItems","SourceItems.Source","SourceItems.Source.MeasurementUnit").
			Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Select(Storage{}.SelectArrayWithoutDataURL())
		})
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,
			[]string{"Images","ProductType","ProductCategories","PaymentSubject","VatCode","MeasurementUnit","Account","ProductCards",
				"Manufacturer","ProductTags","SourceItems","SourceItems.Source","SourceItems.Source.MeasurementUnit"})

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
func (product *Product) AfterCreate(tx *gorm.DB) error {
	// AsyncFire(*Event{}.ProductCreated(product.AccountId, product.Id))
	AsyncFire(NewEvent("ProductCreated", map[string]interface{}{"account_id":product.AccountId, "product_id":product.Id}))

	// Создаем содержание null
	/*if !product.IsKit && product.Id > 0{
		p := Product{}
		err := Account{Id: product.AccountId}.LoadEntity(&p,product.Id, nil)
		if err != nil {
			fmt.Println(err)
			return err
		}

	}*/

	return nil
}

// ######### INTERFACE EVENT Functions ############
// ############# Entity interface #############
func (product Product) GetId() uint { return product.Id }
func (product *Product) setId(id uint) { product.Id = id }
func (product *Product) setPublicId(publicId uint) {product.PublicId = publicId }
func (product Product) GetAccountId() uint { return product.AccountId }
func (product *Product) setAccountId(id uint) { product.AccountId = id }
func (product Product) SystemEntity() bool { return false }
// ############# End of Entity interface #############

// ######### CRUD Functions ############
func (product Product) create() (Entity, error)  {

	_item := product
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}
func (Product) get(id uint, preloads []string) (Entity, error) {
	var item Product

	err := item.GetPreloadDb(false, false, preloads).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
func (product *Product) load(preloads []string) error {
	if product.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить Product - не указан  Id"}
	}

	err := product.GetPreloadDb(false, false, preloads).First(product, product.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (product *Product) loadByPublicId(preloads []string) error {
	if product.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить Product - не указан  Id"}
	}
	if err := product.GetPreloadDb(false,false, preloads).
		First(product, "account_id = ? AND public_id = ?", product.AccountId, product.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (Product) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return Product{}.getPaginationList(accountId, 0,25,sortBy,"",nil,preload)
}
func (Product) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	products := make([]Product,0)
	var total int64

	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&Product{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&products, "label ILIKE ? OR short_label ILIKE ? OR article ILIKE ? OR brand ILIKE ? OR model ILIKE ? OR description ILIKE ?", search,search,search,search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&Product{}).GetPreloadDb(false, false, nil).
			Where("account_id = ? AND label ILIKE ? OR short_label ILIKE ? OR article ILIKE ? OR brand ILIKE ? OR model ILIKE ? OR description ILIKE ?", accountId, search,search,search,search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&Product{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&products).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&Product{}).GetPreloadDb(false, false, nil).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}
	
	entities := make([]Entity, len(products))
	for i := range products {
		entities[i] = &products[i]
	}

	return entities, total, nil
}
func (product *Product) update(input map[string]interface{}, preloads []string) error {
	delete(input,"payment_subject")
	delete(input,"measurement_unit")
	delete(input,"images")
	delete(input,"account")
	delete(input,"product_cards")
	delete(input,"vat_code")
	delete(input,"manufacturer")
	delete(input,"product_type")
	delete(input,"shipments")
	delete(input,"shipment_items")
	delete(input,"warehouse_items")
	delete(input,"warehouses")
	delete(input,"inventories")
	delete(input,"product_categories")
	delete(input,"product_tags")
	delete(input,"sources")
	delete(input,"source_items")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id","payment_subject_id","vat_code_id","measurement_unit_id","manufacturer_id","product_type_id"}); err != nil {
		return err
	}

	if err := product.GetPreloadDb(false, false, nil).Where("id = ?", product.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	AsyncFire(NewEvent("ProductUpdated", map[string]interface{}{"account_id":product.AccountId, "product_id":product.Id}))

	err := product.GetPreloadDb(false,false, preloads).First(product, product.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (product *Product) delete () error {
	if err := product.GetPreloadDb(false,false,nil).Where("id = ?", product.Id).Delete(product).Error;err != nil {
		return err
	}

	AsyncFire(NewEvent("ProductDeleted", map[string]interface{}{"account_id":product.AccountId, "product_id":product.Id}))

	return nil
}
// ######### END CRUD Functions ############

// ########## SELF FUNCTIONAL ############
func (product Product) ExistModel() bool {
	if product.Model == nil {
		return false
	}
	var count int64
	_ = db.Model(&Product{}).Where("account_id = ? AND model = ?", product.AccountId, product.Model).Count(&count)
	if count > 0 {
		return true
	}
	return false
}
func (product Product) AddAttr() error {
	return nil
}
func (product Product) GetAttribute(name string) (interface{}, error) {

	rawData, err := product.Attributes.MarshalJSON()
	if err != nil {
		return "", err
	}

	m := make(map[string]interface{})
	if err = json.Unmarshal(rawData, &m); err != nil {
		return "err", nil
	}

	return m[name], nil
	
}
type PropertyMap map[string]interface{}

func (p PropertyMap) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

func (p *PropertyMap) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}

	var i interface{}
	err := json.Unmarshal(source, &i)
	if err != nil {
		return err
	}

	*p, ok = i.(map[string]interface{})
	if !ok {
		return errors.New("Type assertion .(map[string]interface{}) failed.")
	}

	return nil
}

func (product *Product) AppendProductCategory(productCategory *ProductCategory, strict... bool) error {

	// 1. Загружаем продукт еще раз
	if err := productCategory.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: нельзя добавить категорию, она не найдена"}
	}

	if product.Id < 1 {
		return utils.Error{Message: "Техническая ошибка: нельзя добавить категорию, продукта не загружен"}
	}

	// 2. Проверяем есть ли уже в этой категории этот продукт
	if product.ExistProductCategory(productCategory.Id) {
		if len(strict) > 0 && strict[0] {
			return utils.Error{Message: "Категория уже числиться за товаром"}
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
func (product *Product) RemoveProductCategory(productCategory *ProductCategory) error {

	// 1. Загружаем продукт еще раз
	if err := productCategory.load(nil); err != nil {
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

func (product *Product) SyncProductCategoriesByIds(productCategories []ProductCategory) error {

	// очищаем список категорий
	if product.Id < 1 {
		return utils.Error{Message: "Не найден продукт"}
	}

	if err := db.Model(product).Association("ProductCategories").Clear(); err != nil {
		return err
	}

	for _,_productCategory := range productCategories {
		if err := product.AppendProductCategory(&ProductCategory{Id: _productCategory.Id}, false); err != nil {
			return err
		}
	}

	AsyncFire(NewEvent("ProductSyncProductCategories", map[string]interface{}{"account_id":product.AccountId, "product_id":product.Id}))

	return nil
}

func (product *Product) AppendProductTag(productTag *ProductTag, strict... bool) error {

	// 1. Загружаем продукт еще раз
	if err := productTag.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: нельзя добавить tag, она не найдена"}
	}

	if product.Id < 1 {
		return utils.Error{Message: "Техническая ошибка: нельзя добавить tag, т.к. продукта не загружен"}
	}

	// 2. Проверяем есть ли уже в этой категории этот продукт
	if product.ExistProductTag(productTag.Id) {
		if len(strict) > 0 && strict[0] {
			return utils.Error{Message: "Tag уже числиться за товаром"}
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
func (product *Product) RemoveProductTag(productTag *ProductTag) error {

	// 1. Загружаем продукт еще раз
	if err := productTag.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: нельзя удалить продукт, он не найден"}
	}

	if product.Id < 1 || productTag.Id < 1 {
		return utils.Error{Message: "Техническая ошибка: account id || product id || product tag id == nil"}
	}

	if err := db.Where("product_tag_id = ? AND product_id = ?", productTag.Id, product.Id).Delete(
		&ProductTagProduct{}).Error; err != nil {
		return err
	}

	return nil
}
func (product *Product) SyncProductTagsByIds(ProductTags []ProductTag) error {

	// 1. Загружаем продукт еще раз
	if product.Id < 1 {
		return utils.Error{Message: "Тег не найден"}
	}

	// очищаем список категорий
	if err := db.Model(product).Association("ProductTags").Clear(); err != nil {
		return err
	}

	for _,_productTag := range ProductTags {
		if err := product.AppendProductTag(&ProductTag{Id: _productTag.Id}, false); err != nil {
			return err
		}
	}

	AsyncFire(NewEvent("ProductSyncProductTags", map[string]interface{}{"account_id":product.AccountId, "product_id":product.Id}))

	return nil
}

func (product *Product) ExistProductCategory(productCategoryId uint) bool {

	if productCategoryId < 1 {
		return false
	}

	pcp := ProductCategoryProduct{}
	result := db.Model(&ProductCategoryProduct{}).First(&pcp,"product_category_id = ? AND product_id = ?", productCategoryId, product.Id)

	if result.Error != nil {
		return false
	}
	if result.RowsAffected > 0 {
		return true
	}


	return false
}
func (product *Product) ExistProductTag(tagId uint) bool {

	if tagId < 1 {
		return false
	}

	pcp := ProductTagProduct{}
	if err := db.Model(&ProductTagProduct{}).First(&pcp,"product_tag_id = ? AND product_id = ?", tagId, product.Id).Error; err != nil {
		return false
	}

	return true
}

func (product *Product) AppendSourceItem(source *Product, amountUnits float64, enableViewing bool, strict bool) error {

	// 1. Загружаем продукт-источник еще раз
	if err := source.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: нельзя добавить source, т.к. он не найден"}
	}

	if product.Id < 1 {
		return utils.Error{Message: "Техническая ошибка: нельзя добавить source, т.к. продукта не загружен"}
	}

	if product.IsKit && product.Id == source.Id {
		if strict {
			return utils.Error{Message: "Сборный товар не может состоять из себя же"}
		}
		return nil
	}

	// 2. Проверяем есть ли уже в этой категории этот продукт
	if product.ExistSourceItem(source.Id) {
		if strict {
			return utils.Error{Message: "Source item уже числиться за товаром"}
		} else {

			// update
			if err := db.Model(&ProductSource{}).Where("source_id = ? AND product_id = ?", source.Id, product.Id).
				Updates(map[string]interface{}{"amount_units": amountUnits,"enable_viewing":enableViewing}).Error; err != nil {
				return err
			}
			return nil
		}
	}

	if err := db.Create(
		&ProductSource {
			ProductId: product.Id, SourceId: source.Id, AmountUnits: amountUnits, EnableViewing: enableViewing}).Error; err != nil {
		return err
	}

	return nil
}
func (product *Product) RemoveSourceItem(sourceId uint) error {

	if product.Id < 1 || sourceId < 1 {
		return utils.Error{Message: "Техническая ошибка: product id || source id == nil"}
	}

	if err := db.Where("source_id = ? AND product_id = ?", sourceId, product.Id).Delete(
		&ProductSource{}).Error; err != nil {
		return err
	}
	
	return nil
}
func (product *Product) UpdateSourceItem(sourceId uint, input map[string]interface{}) error {

	if product.Id < 1 || sourceId < 1 {
		return utils.Error{Message: "Техническая ошибка: product id || source id == nil" }
	}

	if err := db.Model(&ProductSource{}).Where("source_id = ? AND product_id = ?", sourceId, product.Id).
		Omit("source_id", "product_id").Updates(input).Error; err != nil {return err}

	/*if err := db.Model().Where("source_id = ? AND product_id = ?", sourceId, product.Id).Updates(input).Error; err != nil {
		return err
	}*/

	return nil
}
func (product *Product) SyncSourceItems(productSources []ProductSource) error {

	// 1. Загружаем продукт еще раз
	if product.Id < 1 {
		return utils.Error{Message: "Товар не найден"}
	}

	// очищаем список связей
	if err := db.Model(product).Association("Sources").Clear(); err != nil {
		return err
	}

	for _,sourceItem := range productSources {
		// fmt.Println("sourceItem.ProductId: ", sourceItem.ProductId)
		// fmt.Println("sourceItem.AmountUnits: ", sourceItem.AmountUnits)
		if product.IsKit && product.Id == sourceItem.ProductId {
			continue
		}
		if err := product.AppendSourceItem( &Product{Id: sourceItem.ProductId}, sourceItem.AmountUnits, sourceItem.EnableViewing, false); err != nil {
			return err
		}
	}

	return nil
}
func (product *Product) ExistSourceItem(sourceId uint) bool {

	if sourceId < 1 {
		return false
	}

	pcp := ProductSource{}
	if err := db.First(&pcp,"source_id = ? AND product_id = ?", sourceId, product.Id).Error; err != nil {
		return false
	}

	return true
}


