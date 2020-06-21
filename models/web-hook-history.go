package models

import "github.com/jinzhu/gorm"

type WebHookHistory struct {
	ID     		uint   	`json:"id" gorm:"primary_key"`
	AccountID 	uint 	`json:"-" gorm:"type:int;index;not null;"`
	WebHookID 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	Status 		bool 	`json:"enabled" gorm:"type:bool;default:true"` // успешен ли вызо
}

func (WebHookHistory) PgSqlCreate() {
	db.CreateTable(&WebHookHistory{})
	db.Model(&WebHookHistory{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}

func (card *ProductCard) BeforeCreate(scope *gorm.Scope) error {
	card.ID = 0
	return nil
}

func (ProductCard) TableName() string {
	return "product_cards"
}