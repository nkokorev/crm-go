package models

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
