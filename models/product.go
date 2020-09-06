package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/nkokorev/crm-go/event"
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
	// Id    		uint   `json:"id" gorm:"primaryKey"`
	Id        	uint 	`json:"id" gorm:"primarykey"`
	// gorm.Model
	PublicId	uint   	`json:"public_id" gorm:"type:int;index;not null;"` // Публичный ID заказа внутри магазина
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // можно ли продавать товар и выводить в карточки
	Name 		string 	`json:"name" gorm:"type:varchar(128);default:''"` // Имя товара, не более 128 символов
	ShortName 	string 	`json:"short_name" gorm:"type:varchar(128);default:''"` // Имя товара, не более 128 символов
	
	Article 	string 	`json:"article" gorm:"type:varchar(128);index;"` // артикул товара из иных соображений (часто публичный)
	SKU 		string 	`json:"sku" gorm:"type:varchar(128);index;"` // уникальный складской идентификатор. 1 SKU = 1 товар (одна модель)
	Model 		string 	`json:"model" gorm:"type:varchar(255);"` // может повторяться для вывода в web-интерфейсе как "одного" товара

	// Base properties
	RetailPrice		float64 `json:"retail_price" gorm:"type:numeric;"` // розничная цена
	WholesalePrice 	float64 `json:"wholesale_price" gorm:"type:numeric;"` // оптовая цена
	PurchasePrice 	float64 `json:"purchase_price" gorm:"type:numeric;"` // закупочная цена
	RetailDiscount 	float64 `json:"retail_discount" gorm:"type:numeric;"` // розничная фактическая скидка

	// Признак предмета расчета
	PaymentSubjectId	uint	`json:"payment_subject_id" gorm:"type:int;not null;"`// товар или услуга ? [вид номенклатуры]
	PaymentSubject 		PaymentSubject `json:"payment_subject"`

	VatCodeId	uint	`json:"vat_code_id" gorm:"type:int;not null;default:1;"`// товар или услуга ? [вид номенклатуры]
	VatCode		VatCode	`json:"vat_code"`

	UnitMeasurementId 		uint	`json:"unit_measurement_id" gorm:"type:int;default:1;"` // тип измерения
	UnitMeasurement 		UnitMeasurement // Ед. измерения: штуки, коробки, комплекты, кг, гр, пог.м.
	
	ShortDescription 	string 	`json:"short_description" gorm:"type:varchar(255);"` // pgsql: varchar - это зачем?)
	Description 		string 	`json:"description" gorm:"type:text;"` // pgsql: text

	// Обновлять только через AppendImage
	// Images 			[]Storage 	`json:"images" gorm:"polymorphic:Owner;"`  // association_autoupdate:false;
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

	Account Account `json:"-"`
	// ProductGroups []ProductGroup `json:"-" gorm:"many2many:product_group_products"`
	ProductCards []ProductCard `json:"product_cards" gorm:"many2many:product_card_products;ForeignKey:id;References:id;"`
}

func (Product) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	if err := db.Migrator().AutoMigrate(&Product{}); err != nil {log.Fatal(err)}
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

// ######### INTERFACE EVENT Functions ############
func (product Product) GetId() uint {
	return product.Id
}
// ######### END OF INTERFAe Functions ############

// ######### CRUD Functions ############
func (product Product) create() (*Product, error)  {
	var newProduct = product
	if err := db.Create(&newProduct).Preload("VatCode").Preload("PaymentSubject").First(&newProduct).Error; err != nil { return nil, err }

	if err := db.Where("id = ?", newProduct.Id).First(&newProduct).Error;err != nil {
		return nil, err
	}

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
	}).
		First(&product, id).Error; err != nil {
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
func (product *Product) update(input map[string]interface{}, preloads []string) error {
	// Приводим в опрядок
	// input = utils.FixJSONB_String(input, []string{"attributes"})

	delete(input, "PaymentSubject")
	delete(input, "UnitMeasurement")
	delete(input, "Images")
	delete(input, "ProductCards")

	input = utils.FixJSONB_MapString(input, []string{"attributes"})
	
	if err := db.Set("gorm:association_autoupdate", false).
		Model(&Product{}).Where("id = ?", product.Id).Omit("id", "account_id").Updates(input).Error; err != nil {
		return err
	}

	event.AsyncFire(Event{}.ProductUpdated(product.AccountId, product.Id))

	return nil
}
func (product *Product) delete () error {
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
func (account Account) GetProductByPublicId(publicId uint) (*Product, error) {
	var product Product
	err := db.Model(&product).First(&product, "public_id = ?", publicId).Error
	if err != nil {
		return nil, err
	}

	if account.Id != product.AccountId {
		return nil, utils.Error{Message: "Товар принадлежит другому аккаунту"}
	}

	return &product, nil
}
func (account Account) GetProductListPagination(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{}) ([]Product, int64, error) {

	products := make([]Product,0)
	var total int64

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
			Limit(limit).Offset(offset).Order(sortBy).
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
			Limit(limit).Offset(offset).Order(sortBy).
			Find(&products, "account_id = ?", account.Id).Error

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
	product.Attributes = jsonInput

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