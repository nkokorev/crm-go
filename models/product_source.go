package models

import (
	"gorm.io/gorm"
	"log"
)

// HELPER FOR M<>M IN PgSQL
type ProductSource struct {

	ProductId  	uint `json:"product_id" gorm:"type:int;index;not null;"`

	// Источник, source
	SourceId 	uint `json:"source_id" gorm:"type:int;index;not null;"`

	// Сколько ед. в одном товаре ()
	Quantity 	float64 `json:"quantity" gorm:"type:numeric;"`
	// AmountUnits 	float64 `json:"amount_units" gorm:"type:numeric;"`

	// Отображать или нет в списке содержание
	EnableViewing	bool 	`json:"enable_viewing" gorm:"type:bool;default:true"`

	Source		Product `json:"source"`
}

func (ProductSource) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&ProductSource{}); err != nil { log.Fatal(err) }
	err := db.Exec("ALTER TABLE product_sources \n    ADD CONSTRAINT product_sources_product_id_fkey FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT product_sources_source_id_fkey FOREIGN KEY (source_id) REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    DROP CONSTRAINT IF EXISTS fk_product_sources_product;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

}

func (ProductSource) BeforeCreate(db *gorm.DB) error {
	return nil
}