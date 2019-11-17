package models

// наши клиенты (заказчики)
type Customer struct {
	ID        	uint 	`gorm:"primary_key;unique_index;" json:"-"` // внутренний ключ, не должен экспортироваться
	HashID 		string 	`json:"id" gorm:"type:varchar(10);unique_index;not null;"` // публичный уникальный ключ категории
	AccountID 	uint 	`json:"-" gorm:"default:null;"` // аккаунт, к которой принадлежит

	CustomerGroups []CustomerGroup `json:"customer_groups" gorm:"many2many:customer_customer_group"`
}

// группы заказчиков
type CustomerGroup struct {
	ID        	uint 	`gorm:"primary_key;unique_index;" json:"-"` // внутренний ключ, не должен экспортироваться
	HashID 		string 	`json:"id" gorm:"type:varchar(10);unique_index;not null;"` // публичный уникальный ключ категории
	AccountID 	uint 	`json:"-" gorm:"default:null;"` // аккаунт, к которой принадлежит
	System 		bool 	`json:"system" gorm:"default:false;"` // есть системные группы: клиенты, поставщики

	Customers []Customer `json:"customers" gorm:"many2many:customer_customer_group"`
}
