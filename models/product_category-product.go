package models

import (
	"gorm.io/gorm"
	"log"
)

// HELPER FOR M<>M IN PgSQL
type ProductCategoryProduct struct {
	ProductId  		uint `json:"product_id"`
	ProductCategoryId 	uint `json:"product_category_id"`
}

func (ProductCategoryProduct) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&ProductCategoryProduct{}); err != nil { log.Fatal(err) }
	err := db.Exec("ALTER TABLE product_category_products \n    ADD CONSTRAINT product_category_products_product_id_fkey FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT product_category_products_product_category_id_fkey FOREIGN KEY (product_category_id) REFERENCES product_categories(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    DROP CONSTRAINT IF EXISTS fk_product_category_productss_product,\n    DROP CONSTRAINT IF EXISTS fk_product_category_products_product_categories;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

}

func (ProductCategoryProduct) BeforeCreate(db *gorm.DB) error {
	return nil
}