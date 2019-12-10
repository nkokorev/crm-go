package models

type OfferProduct struct {
	ID uint	`json:"-"`
	AccountID uint `json:"-"`
	ProductID uint `json:"product_id"`
	OfferID uint `json:"-"`

	Volume float64 `json:"volume"` // сколько списывать со склада

	Account Account `json:"-"`
	Product Product `json:"-"`
	Offer Offer `json:"-"`
	//Product Product	`json:"-"`
}
