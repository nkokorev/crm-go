package models

type Store struct {
	ID        uint `gorm:"primary_key;unique_index;" json:"-"`
	HashID string `json:"hash_id" gorm:"type:varchar(10);unique_index;"`
	AccountID uint `json:"-" gorm:"index;"`
	Name string `json:"name" gorm:"size:255"` // Склад на Маяковке, Склад Дома, Склад на ш. Энтузистов
	//Resource Resource `gorm:"polymorphic:Owner;"`
	//OwnerID   uint
	//OwnerType string //Store, Shop, ...

}
