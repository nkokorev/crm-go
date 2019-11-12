package models

type EntityAttr struct {
	ID        	uint 	`gorm:"primary_key;unique_index;" json:"-"`
	HashID 		string 	`json:"hash_id" gorm:"type:varchar(10);unique_index;not null;"`
	Name 		string 	`json:"name" gorm:"not null;"`
	AccountID 	uint 	`json:"-"`
}
