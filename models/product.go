package models

import (
	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
)

type Product struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"-" gorm:"type:int;index;not_null;"`

	// todo many2many . Один и тот же продукт может быть в разных группах для одного и нескольких магазинов
	// ProductGroupID uint `json:"productGroupId" gorm:"type:int;index;default:null;"` // группа товаров, категория товаров

	// Article string `json:"article"` // артикул товара из иных соображений (часто публичный)
	SKU string `json:"sku" gorm:"default:NULL"` // складской идентификатор
	URL string `json:"url"` // идентификатор страницы (products/syao-chzhun )

	Name string `json:"name"` // Имя товара
	ShortDescription string `json:"shortDescription" gorm:"type:varchar(255);"` // pgsql: varchar
	Description string `json:"description" gorm:"type:text;"` // pgsql: text

	// Specifications Specifications // характеристики товара... (производитель, бренд и т.д. и т.п.)
	// Reviews []Review // Product reviews (отзывы на товар - с рейтингом(?))
	// Questions []question // вопросы по товару
	// Video []Video // видеообзоры по товару на ютубе

	Account Account `json:"-" sql:"-"`
	ProductGroups []ProductGroup `json:"-" gorm:"many2many:product_group_products"`
	// Offers  []Offer `json:"offers" gorm:"many2many:offer_compositions"`
}

func (Product) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&Product{})
	db.Exec("ALTER TABLE products\n    ADD CONSTRAINT products_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\n--     ADD CONSTRAINT products_product_group_id_fkey FOREIGN KEY (product_group_id) REFERENCES product_groups(id) ON DELETE CASCADE ON UPDATE CASCADE;\n")
}

func (product *Product) BeforeCreate(scope *gorm.Scope) error {
	product.ID = 0
	return nil
}

// ######### CRUD Functions ############
func (input Product) create() (*Product, error)  {
	var product = input
	err := db.Create(product).Error
	return &product, err
}

func (Product) get(id uint) (*Product, error) {

	product := Product{}

	if err := db.Table("products").Preload("ProductGroups").First(&product, id).Error; err != nil {
		return nil, err
	}

	return &product, nil
}

func (product *Product) update(input interface{}) error {
	return db.Model(product).Omit("ID", "AccountID").Updates(structs.Map(input)).Error

}

func (product Product) delete () error {
	return db.Model(Product{}).Where("id = ?", product.ID).Delete(product).Error
}
// ######### END CRUD Functions ############