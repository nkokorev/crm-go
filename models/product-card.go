package models

import (
	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

// Карточка "товара" в магазине в котором могут быть разные торговые предложения
type ProductCard struct {
	ID     				uint   `json:"id" gorm:"primary_key"`
	AccountID 			uint `json:"-" gorm:"type:int;index;not_null;"` // потребуется, если productGroupId == null
	// ProductGroupID 		uint `json:"productGroupId" gorm:"type:int;index;default:null;"` // группа товаров, категория товаров

	URL 				string `json:"url" gorm:"type:varchar(255);"` // идентификатор страницы (products/syao-chzhun )
	Breadcrumb 			string `json:"breadcrumb" gorm:"type:varchar(255);default:null;"`
	MetaTitle 			string `json:"meta_title" gorm:"type:varchar(255);default:null;"`
	MetaKeywords 		string `json:"meta_keywords" gorm:"type:varchar(255);default:null;"`
	MetaDescription 	string `json:"meta_description" gorm:"type:varchar(255);default:null;"`

	// Full description нет т.к. в карточке описание берется от офера
	ShortDescription 	string `json:"short_description" gorm:"type:varchar(255);default:null;"` // для превью карточки товара

	// Хелперы карточки: переключение по цветам, размерам и т.д.
	SwitchProducts	 	pq.StringArray `json:"switchProducts" sql:"type:varchar(255)[];default:'{}'"` // {color, size} Параметры переключения среди предложений

	ProductGroups 		[]ProductGroup `json:"productGroups" gorm:"many2many:product_group_product_cards"` // для разделов new и т.д.
	Products 			[]Product `json:"products" gorm:"many2many:product_card_products"` // можно заводить два схожих продукта, разных по цвету
}

func (ProductCard) PgSqlCreate() {
	db.CreateTable(&ProductCard{})
	db.Exec("ALTER TABLE product_cards\n    ADD CONSTRAINT product_cards_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\n--     ADD CONSTRAINT product_cards_product_group_id_fkey FOREIGN KEY (product_group_id) REFERENCES product_groups(id) ON DELETE CASCADE ON UPDATE CASCADE;\n")
}

func (card *ProductCard) BeforeCreate(scope *gorm.Scope) error {
	card.ID = 0
	return nil
}

func (ProductCard) TableName() string {
	return "product_cards"
}

// ######### CRUD Functions ############
func (input ProductCard) create() (*ProductCard, error)  {
	var card = input
	err := db.Create(&card).Error
	return &card, err
}

func (ProductCard) get(id uint) (*ProductCard, error) {

	card := ProductCard{}

	// if err := db.Table(ProductGroup{}.TableName()).Preload("Shop").First(&card, id).Error; err != nil {
	if err := db.First(&card, id).Error; err != nil {
		return nil, err
	}

	return &card, nil
}

func (card *ProductCard) update(input interface{}) error {
	return db.Model(card).Omit("id", "account_id").Updates(structs.Map(input)).Error

}

func (card ProductCard) delete () error {
	return db.Model(ProductCard{}).Where("id = ?", card.ID).Delete(card).Error
}
// ######### END CRUD Functions ############

// ######### SHOP PRODUCT Functions ############
func (shop Shop) CreateProductCard(input ProductCard, group *ProductGroup) (*ProductCard, error) {
	input.AccountID = shop.AccountID
	card, err := input.create()
	if err != nil {
		return nil, err
	}

	if err = group.CreateAndAppendProductCard(card); err != nil {
		return nil, err
	}

	return card, nil
}

// ######### END IF SHOP PRODUCT Functions ############
////// ########

func (card ProductCard) AppendProduct(product *Product) error {
	return db.Model(&card).Association("Products").Append(product).Error
}