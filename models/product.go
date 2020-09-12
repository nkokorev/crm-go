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
	"time"
)

/*
Продукт - как единица товара или услуги. То что потом в чеке у пользователя.
Продукт может быть как шт., упак., так и сборным из других товаров.
Продукт может входить во множество web-карточек (витрин)
Список характеристик продукта не регламентируются, но удобно, когда он принадлежит какой-то группе с фикс. списком параметров.

*/

type Product struct {
	// Id    		uint   `json:"id" gorm:"primaryKey"`
	Id        	uint 	`json:"id" gorm:"primaryKey"`
	// gorm.Model
	PublicId	uint   	`json:"public_id" gorm:"type:int;index;not null;"` // Публичный ID заказа внутри магазина
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	// артикул товара
	Article 	string 	`json:"article" gorm:"type:varchar(128);index;"`

	// Доступен ли товар для продажи в розницу
	RetailSale 	bool 	`json:"retail_sale" gorm:"type:bool;default:true"`

	// Доступен ли товар для продажи оптом
	WholesaleSale	bool	`json:"wholesale_sale" gorm:"type:bool;default:true"`

	// enabled - deprecated
	// Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // можно ли продавать товар и выводить в карточки

	// Этикетка товара
	Label 		string 	`json:"label" gorm:"type:varchar(128);"`

	// торговая марка (Объект!)
	Trademark 	*string	`json:"trademark" gorm:"type:varchar(128);"`

	// Маркировка товара
	Brand 		*string	`json:"brand" gorm:"type:varchar(128);"`

	// mb deprecated
	Name 		string 	`json:"name" gorm:"type:varchar(128);default:''"` // Имя товара, не более 128 символов
	ShortName 	string 	`json:"short_name" gorm:"type:varchar(128);default:''"` // Имя товара, не более 128 символов


	// deprecated, т.к. это относится к складу, а не все товары на складе (есть сбоные и услуги)
	SKU 		string 	`json:"sku" gorm:"type:varchar(128);index;"`

	// Общая тема типа группы товаров, может повторяться для вывода в web-интерфейсе как "одного" товара
	Model 		string 	`json:"model" gorm:"type:varchar(255);"`

	// Base properties
	RetailPrice		float64 `json:"retail_price" gorm:"type:numeric;"` 		// розничная цена
	WholesalePrice 	float64 `json:"wholesale_price" gorm:"type:numeric;"` 	// оптовая цена
	PurchasePrice 	float64 `json:"purchase_price" gorm:"type:numeric;"` 	// закупочная цена
	RetailDiscount 	float64 `json:"retail_discount" gorm:"type:numeric;"` 	// розничная фактическая скидка

	// Вид номенклатуры - ассортиментные группы продаваемых товаров. Привязываются к карточкам..

	// Товарная группа для назначения характеристик
	// PaymentGroupId	uint	`json:"payment_group_id" gorm:"type:int;"`
	// ProductGroup	ProductGroup `json:"product_group"`

	// Тип продукта: улунский, красный (чай), углозачистной станок, шлифовальный станок
	TypeId			*uint	`json:"payment_type_id" gorm:"type:int;"`
	Type			ProductType `json:"product_type"`
	
	// Список продуктов из которых составлен текущий. Это может быть как 1<>1, а может быть и нет (== составной товар)
	WarehouseItems	[]WarehouseItem `json:"warehouse_items"`
	Warehouses		[]Warehouse `json:"warehouses" gorm:"many2many:warehouse_item;"`

	// Сборный ли товар? При нем warehouse_items >= 1. Применяется только к payment_subject = commodity, excise и т.д.
	IsKit			*bool 		`json:"is_kit" gorm:"type:bool;default:false"`

	// Ед. измерения товара: штуки, метры, литры, граммы и т.д.  !!!!
	UnitMeasurementId 		uint	`json:"unit_measurement_id" gorm:"type:int;default:1;"` // тип измерения
	UnitMeasurement 		UnitMeasurement `json:"unit_measurement"`// Ед. измерения: штуки, коробки, комплекты, кг, гр, пог.м.

	// Основные атрибуты для расчета (Можно и в атрибуты)
	Length 	*float64 `json:"length" gorm:"type:numeric;"`
	Width 	*float64 `json:"width" gorm:"type:numeric;"`
	Height 	*float64 `json:"height" gorm:"type:numeric;"`
	Weight 	*float64 `json:"weight" gorm:"type:numeric;"`

	// Производитель (не поставщик)
	ManufacturerId	*uint		`json:"manufacturer_id" gorm:"type:int;"`
	Manufacturer	Manufacturer `json:"manufacturer"`

	// Дата изготовления, дата выпуска, дата производства
	ManufactureDate	*time.Time 	`json:"manufacture_date"`

	// Срок годности, срок хранения (?)
	ShelfLife		*time.Time 	`json:"shelf_life"`

	//  == признак предмета расчета - товар, услуга, работа, набор (комплект) = сборный товар
	// Признак предмета расчета (бухучет - № 54-ФЗ)
	PaymentSubjectId	uint	`json:"payment_subject_id" gorm:"type:int;"`
	PaymentSubject 		PaymentSubject `json:"payment_subject"`
	
	// Ставка НДС или учет НДС (бухучет)
	VatCodeId	uint	`json:"vat_code_id" gorm:"type:int;default:1;"`// товар или услуга ? [вид номенклатуры]
	VatCode		VatCode	`json:"vat_code"`

	ShortDescription 	string 	`json:"short_description" gorm:"type:varchar(255);"` // pgsql: varchar - это зачем?)
	Description 		string 	`json:"description" gorm:"type:text;"` // pgsql: text

	// Обновлять только через AppendImage
	Images 			[]Storage 	`json:"images" gorm:"polymorphic:Owner;"`
	
	Attributes 	datatypes.JSON `json:"attributes" gorm:"type:JSONB;DEFAULT '{}'::JSONB"`

	// todo: можно изменить и сделать свойства товара
	// ключ для расчета веса продукта
	WeightKey 	string `json:"weight_key" gorm:"type:varchar(32);default:'grossWeight'"`

	// Нужно ли считать вес для расчета доставки у данного продукта
	// ConsiderWeight	bool	`json:"considerWeight" gorm:"type:bool;default:false"`

	// Reviews []Review // Product reviews (отзывы на товар - с рейтингом(?))
	// Questions []question // вопросы по товару
	// Video []Video // видеообзоры по товару на ютубе

	// Список поставок товара, в которых он был
	Shipments 		[]Shipment 	`json:"shipments" gorm:"many2many:shipment_products"`

	// История поставок товара в деталях
	ShipmentProduct []ShipmentProduct 	`json:"shipment_product"`

	Account 		Account 		`json:"-"`
	ProductCards 	[]ProductCard 	`json:"product_cards" gorm:"many2many:product_card_products;ForeignKey:id;References:id;"`
}

func (Product) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	if err := db.Migrator().CreateTable(&Product{}); err != nil {log.Fatal(err)}

	// db.Model(&Product{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE products ADD CONSTRAINT products_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
	db.Exec("create unique index uix_products_account_id_sku ON products (account_id,sku) where (length(sku) > 0);\ncreate unique index uix_products_account_id_model ON products (account_id,model) WHERE (length(model) > 0);\ncreate unique index uix_products_account_id_article ON products (account_id,article) WHERE (length(article) > 0);\n-- create unique index uix_products_account_id_sku ON products (account_id,sku) WHERE sku IS NOT NULL;\n")

	err = db.SetupJoinTable(&Product{}, "ProductCards", &ProductCardProduct{})
	if err != nil {
		log.Fatal(err)
	}

	err = db.SetupJoinTable(&Product{}, "Warehouses", &WarehouseItem{})
	if err != nil {
		log.Fatal(err)
	}

	err = db.SetupJoinTable(&Product{}, "Shipments", &ShipmentProduct{})
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
		return db.Preload("PaymentSubject","VatCode","UnitMeasurement","Account","ProductCards").Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Select(Storage{}.SelectArrayWithoutDataURL())
		})
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"PaymentSubject","VatCode","UnitMeasurement","Account","ProductCards","Images"})

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
	if err := product.GetPreloadDb(false,false, preloads).First(product, "account_id = ? AND public_id = ?", product.AccountId, product.PublicId).Error; err != nil {
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
			Find(&products, "name ILIKE ? OR short_name ILIKE ? OR article ILIKE ? OR sku ILIKE ? OR model ILIKE ? OR description ILIKE ?", search,search,search,search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&Product{}).GetPreloadDb(false, false, nil).
			Where("account_id = ? AND name ILIKE ? OR short_name ILIKE ? OR article ILIKE ? OR sku ILIKE ? OR model ILIKE ? OR description ILIKE ?", accountId, search,search,search,search,search,search).
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
	delete(input,"unit_measurement")
	delete(input,"images")
	delete(input,"account")
	delete(input,"product_cards")
	delete(input,"vat_code")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id","payment_subject_id","vat_code_id","unit_measurement_id"}); err != nil {
		return err
	}

	if err := product.GetPreloadDb(false, false, nil).Where("id = ?", product.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	err := product.GetPreloadDb(false,false, preloads).First(product, product.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (product *Product) delete () error {
	return product.GetPreloadDb(false,false,nil).Where("id = ?", product.Id).Delete(product).Error
}
// ######### END CRUD Functions ############

// ########## SELF FUNCTIONAL ############
func (product Product) ExistSKU() bool {
	if len(product.SKU) < 1 {
		return false
	}
	var count int64
	db.Model(&Product{}).Where("account_id = ? AND sku = ?", product.AccountId, product.SKU).Count(&count)
	if count > 0 {
		return true
	}
	return false
}
func (product Product) ExistModel() bool {
	if len(product.Model) < 1 {
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