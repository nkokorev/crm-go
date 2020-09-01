package models

import "gorm.io/gorm"

// HELPER FOR M<>M IN PgSQL
type ProductCardProduct struct {
	ProductId  		uint
	ProductCardId 	uint
	Order			int `json:"order" gorm:"type:int;default:10;"`
}
func (ProductCardProduct) BeforeCreate(db *gorm.DB) error {
	// ...
	return nil
}

func (productCard ProductCard) ExistProduct(product *Product) bool {
	if product.Id < 1 {
		return false
	}
	var count int64
	db.Model(&ProductCardProduct{}).Where("product_id = ? AND product_card_id = ?", product.Id, productCard.Id).Count(&count)
	if count > 0 {
		return true
	}
	return false
}
