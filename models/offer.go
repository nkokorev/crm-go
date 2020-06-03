package models

import "github.com/jinzhu/gorm"

// Торговое предложение для продукта или группы продуктов
type Offer struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"-" gorm:"type:int;index;not_null;"` // Потребуется, если ID карточки товара null
	ProductCardID uint `json:"productCardId" gorm:"type:int;index;default:null;"` // к какой карточке товара относится

	Name string `json:"label" gorm:"type:varchar(255);default:null;"` // публичное название офера
	Price float64 `json:"price"` // стоимость офера
	Discount float64 `json:"discount"` // Скидка, amount

	// Products []Product             `json:"-" gorm:"many2many:offer_compositions"`
	OfferProducts []OfferProduct `json:"offerProducts" gorm:"many2many:offer_products"`

	Account Account `json:"-"`
	ProductCard ProductCard `json:"productCard"`
}

func (Offer) PgSqlCreate() {
	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&Offer{})
	db.Exec("ALTER TABLE offers\n    ADD CONSTRAINT offers_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT offers_product_card_id_fkey FOREIGN KEY (product_card_id) REFERENCES product_cards(id) ON DELETE CASCADE ON UPDATE CASCADE;\n")
}

func (offer *Offer) BeforeCreate(scope *gorm.Scope) error {
	offer.ID = 0
	return nil
}

func (Offer) TableName() string {
	return "offers"
}