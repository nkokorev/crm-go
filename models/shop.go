package models

type Shop struct {
	ID        uint `gorm:"primary_key;unique_index;" json:"-"`
	HashID string `json:"hash_id" gorm:"type:varchar(10);primary_key;unique_index;"`
	Name string `json:"name" gorm:"size:255"` // Магазин pressandgo.ru, koowheel.ru, ш. Энтузистов (офлайн)
	//OwnerID int // тип магазина
	//OwnerType string // тип магазина
}
