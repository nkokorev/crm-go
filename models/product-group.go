package models

import (
	"errors"
	"github.com/jinzhu/gorm"
)

type ProductGroup struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"-" gorm:"type:int;index;not_null;"`
	ParentID uint `json:"parentId" gorm:"default:NULL"`

	ShopID uint `json:"shopId" gorm:"type:int;index;default:null;"` // магазин, к которому относится данная группа

	Code string `json:"code" gorm:"default:NULL"` // tea, coffe, china
	URL string `json:"url" gorm:"default:NULL"`

	Name string `json:"name" gorm:"default:NULL"` // Чай, кофе, ..

	Breadcrumb string `json:"breadcrumb" gorm:"default:NULL"`
	ShortDescription string `json:"shortDescription" gorm:"default:NULL"`
	Description string `json:"description" gorm:"default:NULL"`

	MetaTitle string `json:"metaTitle"`
	MetaKeywords string `json:"metaKeywords"`
	MetaDescription string `json:"metaDescription"`

	Shop        Shop          `json:"-"`
	ParentGroup *ProductGroup `json:"-"`
}

func (ProductGroup) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&ProductGroup{})
	db.Exec("ALTER TABLE product_groups\n    ADD CONSTRAINT products_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT products_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES product_groups(id) ON DELETE CASCADE ON UPDATE CASCADE;\n")

}

func (group *ProductGroup) BeforeCreate(scope *gorm.Scope) error {
	group.ID = 0
	return nil
}

func (ProductGroup) TableName() string {
	return "product_groups"
}


func (pg *ProductGroup) create () error {

	// чекаем на всякий случай ID аккаунта
	if pg.AccountID < 1 {
		return errors.New("Необходимо указать Account ID")
	}

	// тут чекаем уникальность


	return db.Omit("id").Create(pg).Error
}

// Создает и добавляет продукт в продуктовую группу
func (group ProductGroup) CreateProduct(product Product) (*Product, error) {
	// add account id
	// add group id
	return nil, nil
}
