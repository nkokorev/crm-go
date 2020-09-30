package models

import (
	"log"
)

// HELPER FOR M<>M IN PgSQL
type WebPageProductCategories struct {
	WebPageId  			uint `json:"web_page_id"`
	ProductCategoryId 	uint `json:"product_category_id"`

	WebPage			WebPage 		`json:"web_page"`
	ProductCategory	ProductCategory `json:"product_category"`

	// Порядок отображения карточки на странице (нужно ли)
	Priority 	*int		`json:"priority" gorm:"type:int;default:10;"`
}
func (WebPageProductCategories) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&WebPageProductCategories{}); err != nil { log.Fatal(err) }
	err := db.Exec("ALTER TABLE web_page_product_categories \n    ADD CONSTRAINT web_page_product_categories_web_page_id_fkey FOREIGN KEY (web_page_id) REFERENCES web_pages(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT web_page_product_categories_product_category_id_fkey FOREIGN KEY (product_category_id) REFERENCES product_categories(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    DROP CONSTRAINT IF EXISTS fk_web_page_product_categories_web_page,\n    DROP CONSTRAINT IF EXISTS fk_web_page_product_categories_product_category;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}


