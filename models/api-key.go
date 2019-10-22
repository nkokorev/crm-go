package models

import "time"

type ApiKey struct {
	ID			uint `gorm:"primary_key;unique_index;" json:"-"`
	Token 		string `json:"hash_id" gorm:"unique_index;varchar(32)"`
	AccountID 	uint `json:"account_id" gorm:"index;"` // Owner token, foreignKey !
	UserID 		string `json:"creator_id"` // кто создал, можно потом привязать к hash_id
	Label 		string `json:"name" gorm:"size:255"` // 'Токен для сайта'
	Status 		bool `json:"status"`
	Permissions []Permission `json:"permissions" gorm:"many2many:api_key_permissions;"`
	CreatedAt 	time.Time `json:"created_at"`
	UpdatedAt 	time.Time `json:"updated_at"`
}

