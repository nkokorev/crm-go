package models

import (
	"errors"
	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nkokorev/crm-go/utils"
)

type ProductType = string

const (
	ProductTypeCommodity    ProductType = "commodity"
	ProductTypeService      ProductType = "service"
)
/*
Продукт - как единица товара или услуги. То что потом в чеке у пользователя.
Продукт может быть как шт., упак., так и сборным из других товаров.
Продукт может входить во множество web-карточек (витрин)
Список характеристик продукта не регламентируются, но удобно, когда он принадлежит какой-то группе с фикс. списком параметров.

*/
type Product struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"-" gorm:"type:int;index;not null;"`

	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // можно ли продавать товар и выводить в карточки
	Name 		string `json:"name" gorm:"type:varchar(128);default:''"` // Имя товара, не более 128 символов
	ShortName 	string `json:"shortName" gorm:"type:varchar(128);default:''"` // Имя товара, не более 128 символов

	
	Article string `json:"article" gorm:"type:varchar(128);index;default:null"` // артикул товара из иных соображений (часто публичный)
	SKU 	string `json:"sku" gorm:"type:varchar(128);index;default:null;"` // уникальный складской идентификатор. 1 SKU = 1 товар (одна модель)
	Model 	string `json:"model" gorm:"type:varchar(255);default:null"` // может повторяться для вывода в web-интерфейсе как "одного" товара

	// Base properties
	RetailPrice 			float64 `json:"retailPrice" gorm:"type:numeric;default:0"` // розничная цена
	WholesalePrice 			float64 `json:"wholesalePrice" gorm:"type:numeric;default:0"` // оптовая цена
	PurchasePrice 			float64 `json:"purchasePrice" gorm:"type:numeric;default:0"` // закупочная цена
	RetailDiscount 			float64 `json:"retailDiscount" gorm:"type:numeric;default:0"` // розничная фактическая скидка

	ProductType 			ProductType `json:"productType" gorm:"type:varchar(12);default:'commodity';"`// товар или услуга ? [вид номенклатуры]
	UnitMeasurementID 		uint	`json:"unitMeasurementId" gorm:"type:int;default:1;"`
	UnitMeasurement 		UnitMeasurement // Ед. измерения: штуки, коробки, комплекты, кг, гр, пог.м.
	
	// ProductGroupsId uint `json:"productGroupsId"` // группа товара
	// ProductGroups []ProductGroup `json:"productGroups" gorm:"many2many:product_group_products"`

	ShortDescription string `json:"shortDescription" gorm:"type:varchar(255);"` // pgsql: varchar - это зачем?)
	Description 	string `json:"description" gorm:"type:text;"` // pgsql: text

	Images 			[]Storage 	`json:"images" gorm:"PRELOAD:true"`  // ?
	// Attributes []EavAttribute `json:"attributes" gorm:"many2many:product_eav_attributes"` // характеристики товара... (производитель, бренд, цвет, размер и т.д. и т.п.)
	//Attributes []EavAttribute `json:"attributes"` // характеристики товара... (производитель, бренд, цвет, размер и т.д. и т.п.)
	Attributes 		postgres.Jsonb `json:"attributes"`
	// []ProductAttribute // характеристики товара... (производитель, бренд, цвет, размер и т.д. и т.п.)
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
	db.Exec("ALTER TABLE products\n    ADD CONSTRAINT products_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT uix_products_account_id_sku UNIQUE (account_id,sku),\n    ADD CONSTRAINT uix_products_account_id_model UNIQUE (account_id,model),\n    ADD CONSTRAINT uix_products_account_id_article UNIQUE (account_id,article);\n--     ADD CONSTRAINT uix_products_account_id_sku CHECK (account_id AND sku CREATE UNIQUE INDEX ) WHERE sku IS NOT NULL;\n--     ADD constraint uc_products_sku UNIQUE (sku) WHERE sku IS NOT NULL;\n-- ALTER TABLE products ADD CONSTRAINT uc_products_sku UNIQUE (sku);\n\n-- create unique index uix_products_account_id_sku ON products (account_id,sku) WHERE sku IS NOT NULL;\n-- create unique index uix_products_account_id_model ON products (account_id,model) WHERE model IS NOT NULL;\n-- create unique index uix_products_account_id_article ON products (account_id,article) WHERE article IS NOT NULL;\n")
}

func (product *Product) BeforeCreate(scope *gorm.Scope) error {
	product.ID = 0
	return nil
}

// ######### CRUD Functions ############
func (input Product) create() (*Product, error)  {
	var product = input
	err := db.Create(&product).First(&product).Error
	return &product, err
}

func (Product) get(id uint) (*Product, error) {

	product := Product{}

	if err := db.Model(&product).Preload("ProductCards").First(&product, id).Error; err != nil {
		return nil, err
	}

	return &product, nil
}

func (Product) getList(accountId uint) ([]Product, error) {

	products := make([]Product,0)

	err := db.Model(&Product{}).Preload("ProductCards").Find(&products, "account_id = ?", accountId).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return products, nil
}

func (product *Product) update(input interface{}) error {
	return db.Model(product).Omit("id", "account_id").Updates(structs.Map(input)).Error

}

func (product Product) delete () error {
	return db.Model(Product{}).Where("id = ?", product.ID).Delete(product).Error
}
// ######### END CRUD Functions ############

// ######### ACCOUNT Functions ############
func (account Account) CreateProduct(input Product) (*Product, error) {
	input.AccountID = account.ID
	
	if input.ExistSKU() {
		return nil, utils.Error{Message: "Повторение данных", Errors: map[string]interface{}{"sku":"Товар с таким SKU уже есть"}}
	}
	if input.ExistModel() {
		return nil, utils.Error{Message: "Повторение данных", Errors: map[string]interface{}{"model":"Товар с такой моделью уже есть"}}
	}

	product, err := input.create()
	if err != nil {
		return nil, err
	}

	// тут что-то про дефолтный оффер
	/*if group != nil {
		err = group.AppendProduct(product)
		if err != nil {
			return nil, err
		}
	}*/

	return product, nil
}

func (account Account) GetProduct(productId uint) (*Product, error) {
	product, err := Product{}.get(productId)
	if err != nil {
		return nil, err
	}

	if account.ID != product.AccountID {
		return nil, utils.Error{Message: "Товар принадлежит другому аккаунту"}
	}

	return product, nil
}

func (account Account) GetProductListPagination(offset, limit int, search string) ([]Product, uint, error) {

	products := make([]Product,0)

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&Product{}).
			Preload("ProductCards").
			Limit(limit).
			Offset(offset).
			Where("account_id = ?", account.ID).
			Find(&products, "name ILIKE ? OR short_name ILIKE ? OR article ILIKE ? OR sku ILIKE ? OR model ILIKE ? OR description ILIKE ?" , search,search,search,search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

	} else {
		if offset < 0 || limit < 0 {
			return nil, 0, errors.New("Offset or limit is wrong")
		}

		err := db.Model(&Product{}).
			Preload("ProductCards").
			Limit(limit).
			Offset(offset).
			// Joins("LEFT JOIN users ON account_users.user_id = users.id").
			Find(&products, "account_id = ?", account.ID).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}
	}

	// len(cards) != всему списку!
	var total uint
	err := db.Model(&Product{}).Where("account_id = ?", account.ID).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема"}
	}

	return products, total, nil
}

func (account Account) UpdateProduct(productId uint, input interface{}) (*Product, error) {
	product, err := account.GetProduct(productId)
	if err != nil {
		return nil, err
	}

	if account.ID != product.AccountID {
		return nil, utils.Error{Message: "Товар принадлежит другому аккаунту"}
	}

	err = product.update(input)

	return product, err

}

func (account Account) DeleteProduct(productId uint) error {

	// включает в себя проверку принадлежности к аккаунту
	product, err := account.GetProduct(productId)
	if err != nil {
		return err
	}

	return product.delete()
}
// ######### END OF ACCOUNT Functions ############


// ########## SELF FUNCTIONAL ############
func (product Product) ExistSKU() bool {
	return !db.Unscoped().First(&Product{},"account_id = ? AND sku = ?", product.AccountID, product.SKU).RecordNotFound()
}

func (product Product) ExistModel() bool {
	return !db.Unscoped().First(&Product{},"account_id = ? AND model = ?", product.AccountID, product.Model).RecordNotFound()
}

func (product Product) AddAttr() error {
	return nil
}