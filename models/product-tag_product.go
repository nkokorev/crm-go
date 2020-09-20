package models

import (
	"log"
)

// HELPER FOR M<>M IN PgSQL
type ProductTagProduct struct {
	ProductId  		uint 	`json:"product_id"`
	ProductTagId	uint 	`json:"product_tag_id"`
}

func (ProductTagProduct) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&ProductTagProduct{}); err != nil { log.Fatal(err) }
	err := db.Exec("ALTER TABLE product_tag_products \n    ADD CONSTRAINT product_tag_products_id_fkey FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT product_tag_products_product_tag_id_fkey FOREIGN KEY (product_tag_id) REFERENCES product_tags(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    DROP CONSTRAINT IF EXISTS fk_product_tag_products_product_tag,\n    DROP CONSTRAINT IF EXISTS fk_product_tag_products_product,\n    DROP CONSTRAINT IF EXISTS fk_product_tag_products_tag;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}

