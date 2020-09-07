package models

import (
	"gorm.io/gorm"
	"log"
)

// HELPER FOR M<>M IN PgSQL
type ProductCardProduct struct {
	ProductId  		uint `json:"product_id"`
	ProductCardId 	uint `json:"product_card_id"`
	// Order			int `json:"order" gorm:"type:int;default:10;"`
	Priority 	int		`json:"priority" gorm:"type:int;default:10;"` // Порядок отображения (часто нужно файлам)
}

func (ProductCardProduct) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&ProductCardProduct{}); err != nil { log.Fatal(err) }
	err := db.Exec("ALTER TABLE product_card_products \n    ADD CONSTRAINT product_cards_product_id_fkey FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT product_cards_product_card_id_fkey FOREIGN KEY (product_card_id) REFERENCES product_cards(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    DROP CONSTRAINT IF EXISTS fk_product_card_products_product,\n    DROP CONSTRAINT IF EXISTS fk_product_card_products_product_card;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

	err = db.SetupJoinTable(&ProductCard{}, "Products", &ProductCardProduct{})
	if err != nil {
		log.Fatal(err)
	}
	err = db.SetupJoinTable(&ProductCard{}, "WebPages", &WebPageProductCard{})
	if err != nil {
		log.Fatal(err)
	}
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
