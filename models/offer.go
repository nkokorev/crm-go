package models


type Offer struct {
	ID uint	`json:"id"`
	AccountID uint `json:"-"`

	Name string `json:"label"`
	Price float64 `json:"price"`

	//EavAttributeID uint `json:"-"`
	//Properties []EavProductOfferAttribute `json:"properties"`
	//Properties []Property `json:"properties"`

	Volume float64 `json:"volume"`

	Products []Product `json:"products" gorm:"many2many:product_product_offer"`

	Account Account `json:"-"`
	//Product Product	`json:"-"`
}
