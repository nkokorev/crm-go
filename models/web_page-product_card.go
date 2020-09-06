package models

import "gorm.io/gorm"

// HELPER FOR M<>M IN PgSQL
type WebPageProductCard struct {
	WebPageId  		uint
	ProductCardId 	uint

	// Порядок отображения карточки на странице (нужно ли)
	Priority 	int		`json:"priority" gorm:"type:int;default:10;"`
}

func (WebPageProductCard) BeforeCreate(db *gorm.DB) error {
	return nil
}

func (webPage WebPage) ExistProductCard(productCard *ProductCard) bool {

	if productCard.Id < 1 {
		return false
	}

	var count int64

	db.Model(&WebPageProductCard{}).Where("web_page_id = ? AND product_card_id = ?", webPage.Id, productCard.Id).Count(&count)

	if count > 0 {
		return true
	}

	return false
}
