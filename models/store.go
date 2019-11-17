package models

// Магазин,
type Store struct {
	ID        	uint 	`gorm:"primary_key;unique_index;" json:"-"` // внутренний ключ, не должен экспортироваться
	HashID 		string 	`json:"id" gorm:"type:varchar(10);unique_index;not null;"` // публичный уникальный ключ категории (store_id)
	AccountID 	uint `json:"-" gorm:"index;"`
	Name 		string `json:"name" gorm:"size:255"` // Склад на Маяковке, Склад Дома, Склад на ш. Энтузистов
	WebSiteID	uint	`json:"web_site_id" gorm:"not null"`
}
