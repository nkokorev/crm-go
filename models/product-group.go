package models

import (
	"errors"
)

type ProductGroup struct {
	ID uint	`json:"id"`
	AccountID uint `json:"-"`
	ParentID uint `json:"parent_id" gorm:"default:NULL"`

	ShopID uint `json:"shop_id"` // магазин, к которому относится данная группа

	Code string `json:"code" gorm:"default:NULL"` // tea, coffe, china
	URL string `json:"url" gorm:"default:NULL"`

	Name string `json:"name" gorm:"default:NULL"` // Чай, кофе, ..

	Breadcrumb string `json:"breadcrumb" gorm:"default:NULL"`
	ShortDescription string `json:"short_description" gorm:"default:NULL"`
	Description string `json:"description" gorm:"default:NULL"`

	MetaTitle string `json:"meta_title"`
	MetaKeywords string `json:"meta_keywords"`
	MetaDescription string `json:"meta_description"`

	Shop Shop `json:"-"`
	ParentGroup *ProductGroup `json:"-"`
}

func (pg *ProductGroup) Create () error {

	// чекаем на всякий случай ID аккаунта
	if pg.AccountID < 1 {
		return errors.New("Необходимо указать Account ID")
	}

	// тут чекаем уникальность


	return db.Omit("id").Create(pg).Error
}
