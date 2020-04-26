package models

import "time"

type Order struct {
	ID uint	`json:"id"`
	AccountID uint `json:"-"`


	User User	`json:"user"`
	Products []Product `json:"-" gorm:"many2many:offer_compositions"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt *time.Time `json:"-" sql:"index"`
}
