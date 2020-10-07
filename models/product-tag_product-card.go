package models

import (
	"log"
)

// HELPER FOR M<>M IN PgSQL
type ProductTagProductCard struct {
	ProductCardId  	uint 	`json:"product_card_id" gorm:"type:int;index;not null;"`
	ProductTagId	uint 	`json:"product_tag_id" gorm:"type:int;index;not null;"`
}

func (ProductTagProductCard) PgSqlCreate() {
	if !db.Migrator().HasTable(&ProductTagProductCard{}) {
		if err := db.Migrator().CreateTable(&ProductTagProductCard{}); err != nil { log.Fatal(err) }
		err := db.Exec("ALTER TABLE product_tag_product_cards \n    ADD CONSTRAINT product_tag_product_cards_product_card_id_fkey FOREIGN KEY (product_card_id) REFERENCES product_cards(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT product_tag_product_cards_product_tag_id_fkey FOREIGN KEY (product_tag_id) REFERENCES product_tags(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    DROP CONSTRAINT IF EXISTS fk_product_tag_product_cards_product_tag,\n    DROP CONSTRAINT IF EXISTS fk_product_tag_product_cards_product,\n    DROP CONSTRAINT IF EXISTS fk_product_tag_product_cards_tag;\n\ncreate unique index uix_product_tag_product_cards_product_card_id_product_tag_id ON product_tag_product_cards (product_card_id,product_tag_id);").Error
		if err != nil {
			log.Fatal("Error: ", err)
		}
	}

}

