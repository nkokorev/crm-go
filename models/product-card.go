package models

import "errors"

type ProductCard struct {
	ID uint	`json:"id"`
	AccountID uint `json:"-"`
	ShopID uint `json:"shop_id"`

	URL string `json:"url"` // идентификатор страницы (products/syao-chzhun )
	Breadcrumb string `json:"breadcrumb" gorm:"default:NULL"`
	ShortDescription string `json:"short_description" gorm:"default:NULL"`
	Description string `json:"description" gorm:"default:NULL"`

	MetaTitle string `json:"meta_title"`
	MetaKeywords string `json:"meta_keywords"`
	MetaDescription string `json:"meta_description"`
	
	//Products []Product `json:"products" gorm:"many2many:product_product_card"`
	Offers []Offer `json:"offers" gorm:"many2many:product_card_offers"`

	Shop Shop `json:"-"`
}

// M:M ProductCard <> Offers
type ProductCardOffers struct {
	AccountID,OfferID,ProductCardID uint
	Order int
}

func (pc *ProductCard) Create () error {

	// чекаем на всякий случай ID аккаунта
	if pc.AccountID < 1 {
		return errors.New("Необходимо указать Account ID")
	}

	return db.Omit("id").Create(pc).Error
}

func (pc *ProductCard) OfferAppend (offer Offer, opt_v... int) error {
	order := 1;
	if len(opt_v) > 0 {
		order = opt_v[0]
	}
	return db.Model(ProductCardOffers{}).Create(&ProductCardOffers{AccountID:pc.AccountID, OfferID:offer.ID, ProductCardID:pc.ID, Order:order}).Error
}

func (pc ProductCard) GetAll(v_opt... uint) (pcs []ProductCard, err error) {

	account_id := pc.AccountID
	if len(v_opt) > 0 {
		account_id = v_opt[0]
	}
	err = db.Model(ProductCard{}).Order("id asc").Preload("Offers.Composition").Where("account_id = ?", account_id).Find(&pcs).Error
	return
}
