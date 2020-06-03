package models

import "github.com/jinzhu/gorm"

// Количество каждого продукта в офере  (Offer Product Composition)
type OfferProduct struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	ProductID uint `json:"productId" gorm:"type:int;index;not_null;"`
	OfferID uint `json:"offerId" gorm:"type:int;index;not_null;"`

	Volume float64 `json:"volume"` // сколько списывать со склада

	// Account Account `json:"-"`
	Product Product `json:"-" gorm:"preload:true"`
	Offer   Offer   `json:"-" gorm:"preload:true"`
}

func (OfferProduct) PgSqlCreate() {
	// 1. Создаем таблицу и настройки в pgSql
	db.DropTableIfExists(&OfferProduct{})
	db.CreateTable(&OfferProduct{})
	db.Exec("ALTER TABLE offer_products\n    ADD CONSTRAINT offer_products_product_id_fkey FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT offer_products_offer_id_fkey FOREIGN KEY (offer_id) REFERENCES offers(id) ON DELETE CASCADE ON UPDATE CASCADE;\n")
}

func (op *OfferProduct) BeforeCreate(scope *gorm.Scope) error {
	op.ID = 0
	return nil
}

func (OfferProduct) TableName() string {
	return "offer_products"
}