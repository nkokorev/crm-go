package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"log"
)

/*
Продукт - как единица товара или услуги. То что потом в чеке у пользователя.
Продукт может быть как шт., упак., так и сборным из других товаров.
Продукт может входить во множество web-карточек (витрин)
Список характеристик продукта не регламентируются, но удобно, когда он принадлежит какой-то группе с фикс. списком параметров.

*/

type Product struct {
	Id     uint   `json:"id" gorm:"primary_key"`
	AccountId uint `json:"-" gorm:"type:int;index;not null;"`

	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // можно ли продавать товар и выводить в карточки
	Name 		string `json:"name" gorm:"type:varchar(128);default:''"` // Имя товара, не более 128 символов
	ShortName 	string `json:"shortName" gorm:"type:varchar(128);default:''"` // Имя товара, не более 128 символов

	
	Article *string `json:"article" gorm:"type:varchar(128);index;default:NULL"` // артикул товара из иных соображений (часто публичный)
	SKU 	*string `json:"sku" gorm:"type:varchar(128);index;default:NULL;"` // уникальный складской идентификатор. 1 SKU = 1 товар (одна модель)
	Model 	*string `json:"model" gorm:"type:varchar(255);default:NULL"` // может повторяться для вывода в web-интерфейсе как "одного" товара

	// Base properties
	RetailPrice 			float64 `json:"retailPrice" gorm:"type:numeric;default:0"` // розничная цена
	WholesalePrice 			float64 `json:"wholesalePrice" gorm:"type:numeric;default:0"` // оптовая цена
	PurchasePrice 			float64 `json:"purchasePrice" gorm:"type:numeric;default:0"` // закупочная цена
	RetailDiscount 			float64 `json:"retailDiscount" gorm:"type:numeric;default:0"` // розничная фактическая скидка

	// Признак предмета расчета
	PaymentSubjectId	uint	`json:"paymentSubjectId" gorm:"type:int;not null;"`// товар или услуга ? [вид номенклатуры]
	PaymentSubject 		PaymentSubject `json:"paymentSubject"`

	VatCodeId	uint	`json:"vatCodeId" gorm:"type:int;not null;default:1;"`// товар или услуга ? [вид номенклатуры]
	VatCode		VatCode	`json:"vatCode"`

	UnitMeasurementId 		uint	`json:"unitMeasurementId" gorm:"type:int;default:1;"` // тип измерения
	UnitMeasurement 		UnitMeasurement // Ед. измерения: штуки, коробки, комплекты, кг, гр, пог.м.
	
	ShortDescription string 	`json:"shortDescription" gorm:"type:varchar(255);"` // pgsql: varchar - это зачем?)
	Description 	string 		`json:"description" gorm:"type:text;"` // pgsql: text

	// Обновлять только через AppendImage
	Images 			[]Storage 	`json:"images" gorm:"polymorphic:Owner;"`  // association_autoupdate:false;
	//Image 			Storage 	`json:"images" gorm:"polymorphic:Storage;" sql:"-"`  // gorm:"polymorphic:Owner;"
	// Attributes []EavAttribute `json:"attributes" gorm:"many2many:product_eav_attributes"` // характеристики товара... (производитель, бренд, цвет, размер и т.д. и т.п.)
	//Attributes []EavAttribute `json:"attributes"` // характеристики товара... (производитель, бренд, цвет, размер и т.д. и т.п.)
	Attributes 		postgres.Jsonb `json:"attributes" gorm:"type:JSONB;DEFAULT '{}'::JSONB"`

	// todo: можно изменить и сделать свойства товара
	// ключ для расчета веса продукта
	WeightKey 		string `json:"weightKey" gorm:"type:varchar(32);default:'grossWeight'"`

	// Нужно ли считать вес для расчета доставки у данного продукта
	// ConsiderWeight	bool	`json:"considerWeight" gorm:"type:bool;default:false"`

	// Attributes 		PropertyMap `json:"attributes" gorm:"type:JSONB;DEFAULT '{}'::JSONB"`
	// Reviews []Review // Product reviews (отзывы на товар - с рейтингом(?))
	// Questions []question // вопросы по товару
	// Video []Video // видеообзоры по товару на ютубе

	Account Account `json:"-"`
	// ProductGroups []ProductGroup `json:"-" gorm:"many2many:product_group_products"`
	ProductCards []ProductCard `json:"productCards" gorm:"many2many:product_card_products"`
}

func (Product) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&Product{})
	db.Exec("ALTER TABLE products\n    ADD CONSTRAINT products_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\n--     ADD CONSTRAINT uix_products_account_id_sku UNIQUE (account_id,sku),\n--     ADD CONSTRAINT uix_products_account_id_model UNIQUE (account_id,model),\n--     ADD CONSTRAINT uix_products_account_id_article UNIQUE (account_id,article);\n--     ADD CONSTRAINT uix_products_account_id_sku CHECK (account_id AND sku CREATE UNIQUE INDEX ) WHERE sku IS NOT NULL;\n--     ADD constraint uc_products_sku UNIQUE (sku) WHERE sku IS NOT NULL;\n-- ALTER TABLE products ADD CONSTRAINT uc_products_sku UNIQUE (sku);\n\ncreate unique index uix_products_account_id_sku ON products (account_id,sku) WHERE sku IS NOT NULL;\ncreate unique index uix_products_account_id_model ON products (account_id,model) WHERE model IS NOT NULL;\ncreate unique index uix_products_account_id_article ON products (account_id,article) WHERE article IS NOT NULL;\n")
}

func (product *Product) BeforeCreate(scope *gorm.Scope) error {
	product.Id = 0
	return nil
}

// ######### INTERFACE EVENT Functions ############
func (product Product) GetId() uint {
	return product.Id
}
// ######### END OF INTERFAe Functions ############

// ######### CRUD Functions ############
func (product Product) create() (*Product, error)  {
	var newProduct = product
	if err := db.Create(&newProduct).Preload("VatCode").Preload("PaymentSubject").First(&newProduct).Error; err != nil { return nil, err }

	event.AsyncFire(Event{}.ProductCreated(newProduct.AccountId, newProduct.Id))
	return &newProduct, nil
}

func (Product) get(id uint) (*Product, error) {

	product := Product{}

	//if err := db.Model(&product).Preload("ProductCards").First(&product, id).Error; err != nil {
	//	return nil, err
	//}
	if err := db.Model(&product).Preload("VatCode").Preload("PaymentSubject").Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Select(Storage{}.SelectArrayWithoutDataURL())
	}).First(&product, id).Error; err != nil {
		return nil, err
	}

	return &product, nil
}

func (Product) getList(accountId uint) ([]Product, error) {

	products := make([]Product,0)

	err := db.Model(&Product{}).Preload("PaymentSubject").Preload("ProductCards").Find(&products, "account_id = ?", accountId).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return products, nil
}

func (product *Product) update(input map[string]interface{}) error {
	delete(input, "PaymentSubject")
	delete(input, "UnitMeasurement")
	delete(input, "Images")
	delete(input, "ProductCards")
	
	err := db.Set("gorm:association_autoupdate", false).
		Model(&Product{}).Where("id = ?", product.Id).Omit("id", "account_id").Updates(input).Error
	if err != nil {
		return err
	}

	event.AsyncFire(Event{}.ProductUpdated(product.AccountId, product.Id))

	return nil
}

func (product Product) delete () error {
	if err := db.Model(Product{}).Where("id = ?", product.Id).Delete(product).Error; err != nil {
		return err
	}
	
	event.AsyncFire(Event{}.ProductDeleted(product.AccountId, product.Id))

	return nil
}
// ######### END CRUD Functions ############

// ######### ACCOUNT Functions ############
func (account Account) CreateProduct(input Product) (*Product, error) {
	input.AccountId = account.Id
	
	if input.ExistSKU() {
		return nil, utils.Error{Message: "Повторение данных SKU", Errors: map[string]interface{}{"sku":"Товар с таким SKU уже есть"}}
	}
	if input.ExistModel() {
		return nil, utils.Error{Message: "Повторение данных", Errors: map[string]interface{}{"model":"Товар с такой моделью уже есть"}}
	}

	product, err := input.create()
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (account Account) GetProduct(productId uint) (*Product, error) {
	product, err := Product{}.get(productId)
	if err != nil {
		return nil, err
	}

	if account.Id != product.AccountId {
		return nil, utils.Error{Message: "Товар принадлежит другому аккаунту"}
	}

	return product, nil
}

func (account Account) GetProductListPagination(offset, limit int, search string) ([]Product, uint, error) {

	products := make([]Product,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&Product{}).
			Preload("VatCode").
			Preload("PaymentSubject").
			Preload("ProductCards").
			Preload("Images", func(db *gorm.DB) *gorm.DB {
				return db.Select(Storage{}.SelectArrayWithoutDataURL())
			}).
			Limit(limit).
			Offset(offset).
			Order("id").
			Where("account_id = ?", account.Id).
			Find(&products, "name ILIKE ? OR short_name ILIKE ? OR article ILIKE ? OR sku ILIKE ? OR model ILIKE ? OR description ILIKE ?", search,search,search,search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Product{}).
			Where("account_id = ? AND name ILIKE ? OR short_name ILIKE ? OR article ILIKE ? OR sku ILIKE ? OR model ILIKE ? OR description ILIKE ?", account.Id, search,search,search,search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		if offset < 0 || limit < 0 {
			return nil, 0, errors.New("Offset or limit is wrong")
		}

		err := db.Model(&Product{}).
			Preload("VatCode").
			Preload("PaymentSubject").
			Preload("ProductCards").
			Preload("Images", func(db *gorm.DB) *gorm.DB {
				return db.Select(Storage{}.SelectArrayWithoutDataURL())
			}).
			Limit(limit).Offset(offset).Order("id").Find(&products, "account_id = ?", account.Id).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Product{}).Where("account_id = ?", account.Id).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	return products, total, nil
}

func (account Account) UpdateProduct(productId uint, input map[string]interface{}) (*Product, error) {

	product, err := account.GetProduct(productId)
	if err != nil {
		return nil, err
	}

	if account.Id != product.AccountId {
		return nil, utils.Error{Message: "Товар принадлежит другому аккаунту"}
	}

	// parse attrs
	jsonInput, err := json.Marshal(input["attributes"])
	if err != nil {
		log.Fatal("Eroror json: ", err)
	}
	product.Attributes = postgres.Jsonb{RawMessage: jsonInput}

	err = product.update(input)
	if err != nil {
		return nil, err
	}

	// todo: костыль вместо евента
	// go account.CallWebHookIfExist(EventProductUpdated, product)

	return product, err

}

func (account Account) DeleteProduct(productId uint) error {

	// включает в себя проверку принадлежности к аккаунту
	product, err := account.GetProduct(productId)
	if err != nil {
		return err
	}

	err = product.delete()
	if err !=nil { return err }

	return nil
}
// ######### END OF ACCOUNT Functions ############


// ########## SELF FUNCTIONAL ############
func (product Product) ExistSKU() bool {
	return !db.Unscoped().First(&Product{},"account_id = ? AND sku = ?", product.AccountId, product.SKU).RecordNotFound()
}

func (product Product) ExistModel() bool {
	return !db.Unscoped().First(&Product{},"account_id = ? AND model = ?", product.AccountId, product.Model).RecordNotFound()
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