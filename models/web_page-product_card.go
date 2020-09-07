package models

import (
	"gorm.io/gorm"
	"log"
)

// HELPER FOR M<>M IN PgSQL
type WebPageProductCard struct {
	WebPageId  		uint `json:"web_page_id"`
	ProductCardId 	uint `json:"product_card_id"`

	// Порядок отображения карточки на странице (нужно ли)
	Priority 	int		`json:"priority" gorm:"type:int;default:10;"`
}
func (WebPageProductCard) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&WebPageProductCard{}); err != nil { log.Fatal(err) }
	err := db.Exec("ALTER TABLE web_page_product_cards \n    ADD CONSTRAINT product_cards_web_page_id_fkey FOREIGN KEY (web_page_id) REFERENCES web_pages(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT product_cards_product_card_id_fkey FOREIGN KEY (product_card_id) REFERENCES product_cards(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    DROP CONSTRAINT IF EXISTS fk_web_page_product_cards_product_card,\n    DROP CONSTRAINT IF EXISTS fk_web_page_product_cards_web_page;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
	err = db.SetupJoinTable(&WebPage{}, "ProductCards", &WebPageProductCard{})
	if err != nil {
		log.Fatal(err)
	}
	err = db.SetupJoinTable(&Product{}, "ProductCards", &ProductCardProduct{})
	if err != nil {
		log.Fatal(err)
	}
}

func (WebPageProductCard) BeforeCreate(db *gorm.DB) error {
	return nil
}

func (webPage WebPage) ExistProductCard(productCard *ProductCard) bool {

	if productCard.Id < 1 {
		return false
	}

	var count int64

	db.Model(&WebPageProductCard{}).Where("web_page_id = ? AND product_card_id = ?", webPage.Id, productCard.Id).Count(&count)

	if count > 0 {
		return true
	}

	return false
}
