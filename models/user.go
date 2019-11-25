package models

import (
	"time"
)

type User struct {
	ID        	uint `gorm:"primary_key;unique_index;" json:"-"`
	HashID 		string `json:"hash_id" gorm:"type:varchar(10);unique_index;not null;"`
	Username 	string `json:"username" gorm:"unique_index;not null;"`
	Email 		string `json:"email" gorm:"unique_index;not null;"`
	Name 		string `json:"name"`
	Surname 	string `json:"surname"`
	Patronymic 	string `json:"patronymic"`
	Password 	string `json:"-" gorm:"not null"` // json:"-"
	DefaultAccountHashID string `json:"default_account_hash_id" gorm:"type:varchar(10);default:NULL;"` // указывает какой аккаунт по дефолту загружать
	//AccountID uint `json:"default_account_id" gorm:"foreignkey:AccountID;default:NULL;"`
	CreatedAt *time.Time `json:"created_at;omitempty"`
	UpdatedAt *time.Time `json:"updated_at;omitempty"`
	DeletedAt *time.Time `sql:"index" json:"-"`
}
