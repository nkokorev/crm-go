package models

import "github.com/jinzhu/gorm"

// Карточка "товара" в магазине в котором могут быть разные торговые предложения
type ProductCard struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"-" gorm:"type:int;index;not_null;"` // потребуется, если productGroupId == null
	ProductGroupID uint `json:"productGroupId" gorm:"type:int;index;default:null;"` // группа товаров, категория товаров

	URL string `json:"url" gorm:"type:varchar(255);"` // идентификатор страницы (products/syao-chzhun )
	Breadcrumb string `json:"breadcrumb" gorm:"type:varchar(255);default:null;"`
	ShortDescription string `json:"short_description" gorm:"type:varchar(255);default:null;"`
	Description string `json:"description" gorm:"type:varchar(255);default:null;"`

	MetaTitle string `json:"meta_title" gorm:"type:varchar(255);default:null;"`
	MetaKeywords string `json:"meta_keywords" gorm:"type:varchar(255);default:null;"`
	MetaDescription string `json:"meta_description" gorm:"type:varchar(255);default:null;"`
	
	ProductGroup ProductGroup `json:"productGroup"`
	Offers []Offer `json:"offers"`
}

func (ProductCard) PgSqlCreate() {
	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&ProductCard{})
	db.Exec("ALTER TABLE product_cards\n    ADD CONSTRAINT product_cards_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT product_cards_product_group_id_fkey FOREIGN KEY (product_group_id) REFERENCES product_groups(id) ON DELETE CASCADE ON UPDATE CASCADE;\n")
}

func (card *ProductCard) BeforeCreate(scope *gorm.Scope) error {
	card.ID = 0
	return nil
}

func (ProductCard) TableName() string {
	return "product_cards"
}

