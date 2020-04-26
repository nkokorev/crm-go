package models

import (
	"errors"
)

type Offer struct {
	ID uint	`json:"id" gorm:"primary_key"`
	AccountID uint `json:"-"`

	Name string `json:"label"` // публичное название офера
	Price float64 `json:"price"` // стоимость офера
	Discount float64 `json:"discount"` // Скидка, amount

	Products []Product `json:"-" gorm:"many2many:offer_compositions"`
	Composition []OfferComposition `json:"composition"`

	Account Account `json:"-"`
	Product Product	`json:"-"`
}

func (offer *Offer) Create () error {

	if offer.AccountID < 1 {
		return errors.New("Необходимо указать Account ID")
	}

	return db.Omit("id").Create(offer).Error
}

func (offer *Offer) ProductAppend (product Product, volume float64) error {
	return db.Model(OfferComposition{}).Create(&OfferComposition{AccountID:offer.AccountID, OfferID:offer.ID, ProductID:product.ID, Volume:volume}).Error
}

func (offer Offer) GetAll(v_opt... uint) (offers []Offer, err error) {

	account_id := offer.AccountID
	if len(v_opt) > 0 {
		account_id = v_opt[0]
	}
	err = db.Model(Offer{}).Order("id asc").Preload("Composition").Where("account_id = ?", account_id).Find(&offers).Error
	return
}
