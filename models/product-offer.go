package models

type ProductOffer struct {
	ID uint	`json:"id"`
	AccountID uint `json:"-"`
	ProductID uint `json:"-"`

	SKU string `json:"sku" gorm:"default:NULL"`
	Label string `json:"label" gorm:"default:NULL"`

	//EavAttributeID uint `json:"-"`
	//Properties []EavProductOfferAttribute `json:"properties"`
	//Properties []Property `json:"properties"`

	Account Account `json:"-"`
	Product Product	`json:"-"`
}
