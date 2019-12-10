package models

import (
	"errors"
)

type Offer struct {
	ID uint	`json:"id"`
	AccountID uint `json:"-"`

	Name string `json:"label"`
	Price float64 `json:"price"`
	Discount float64 `json:"discount"` // Скидка, amount

	//EavAttributeID uint `json:"-"`
	//Properties []EavProductOfferAttribute `json:"properties"`
	//Properties []Property `json:"properties"`

	//Volume float64 `json:"volume"`

	//Products []Product `json:"-" gorm:"many2many:product_product_offer"`
	Products []Product `json:"-" gorm:"many2many:offer_products"`
	Composition []OfferProduct `json:"composition"`

	Account Account `json:"-"`
	//Product Product	`json:"-"`
}

func (offer *Offer) Create () error {

	// чекаем на всякий случай ID аккаунта
	if offer.AccountID < 1 {
		return errors.New("Необходимо указать Account ID")
	}

	return db.Omit("id").Create(offer).Error
}

func (offer *Offer) ProductAppend (product Product, volume float64) error {
	return db.Model(OfferProduct{}).Create(&OfferProduct{AccountID:offer.AccountID, OfferID:offer.ID, ProductID:product.ID, Volume:volume}).Error
	//return db.Model(offer).Association("Products").Append(&product).Error
}


