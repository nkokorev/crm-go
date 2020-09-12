package models

import (
	"gorm.io/gorm"
	"log"
)

// HELPER FOR M<>M IN PgSQL
type ProductCategoryProductCard struct {
	ProductCardId  		uint `json:"product_card_id"`
	ProductCategoryId 	uint `json:"product_category_id"`

	// Порядок отображения карточки на странице (нужно ли)
	Priority 	int		`json:"priority" gorm:"type:int;default:10;"`
}
func (ProductCategoryProductCard) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&ProductCategoryProductCard{}); err != nil { log.Fatal(err) }
	err := db.Exec("ALTER TABLE product_category_product_cards \n    ADD CONSTRAINT product_category_product_cards_product_card_id_fkey FOREIGN KEY (product_card_id) REFERENCES product_cards(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT product_category_product_cards_product_category_id_fkey FOREIGN KEY (product_card_id) REFERENCES product_categories(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    DROP CONSTRAINT IF EXISTS fk_product_category_product_cards_product_card,\n    DROP CONSTRAINT IF EXISTS fk_product_category_product_cards_product_category;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}

func (ProductCategoryProductCard) BeforeCreate(db *gorm.DB) error {
	return nil
}

func (webPage WebPage) ExistProductCard(productCard *ProductCard) bool {

	if productCard.Id < 1 {
		return false
	}

	var count int64

	db.Model(&ProductCategoryProductCard{}).Where("web_page_id = ? AND product_card_id = ?", webPage.Id, productCard.Id).Count(&count)

	if count > 0 {
		return true
	}

	return false
}
