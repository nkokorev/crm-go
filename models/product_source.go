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

// Вызывает событие обновления доступных складских запасов для всех товаров имеющих в составе source_id
func (ProductSource) CreateEventUpdateBySourceId(accountId, sourceId uint, updateKitOnly bool )  {

	if accountId < 1 || sourceId < 1 {
		return
	}
	productSources := make([]ProductSource,0)
	// if err := db.Model(&ProductSource{}).Where("source_id = ? AND product_id <> ?", sourceId, sourceId).
	if err := db.Model(&ProductSource{}).Where("source_id = ?", sourceId).
		Find(&productSources).Error; err != nil {return}

	for i := range productSources {

		// Обновление не составного продукта 1 = 1
		if updateKitOnly && (sourceId == productSources[i].ProductId) {
			continue
		}
		AsyncFire(NewEvent("ProductUpdated", map[string]interface{}{"account_id":accountId, "product_id":productSources[i].ProductId}))
	}

}

func (ProductSource) CreateEventUpdateByProductId(accountId, productId uint, updateKitOnly bool )  {

	if accountId < 1 || productId < 1 {
		return
	}

	productSources := make([]ProductSource,0)
	
	if err := db.Model(&ProductSource{}).Where("product_id = ?", productId).
		Find(&productSources).Error; err != nil {
			return
		}

	for i := range productSources {
		ProductSource{}.CreateEventUpdateBySourceId(accountId,productSources[i].SourceId,updateKitOnly)
	}

}