package models

import (
	"fmt"
	"github.com/nkokorev/crm-go/utils"
)

// EAV-Атрибуты, предусмотренные в аккаунте: Размер одежды, Тип упаковки, Состав, Цвет и т.д.
type EavAttribute struct {
	ID uint	`json:"id"`
	AccountID uint `json:"-"`

	Code string	`json:"code"`
	Label string `json:"label"`

	Multiple bool `json:"multiple"`
	Required bool `json:"required"`

	// это данные о хранении и отображении атрибута
	AttrTypeCode string `json:"attr_type_code"` // inline, text_field, text_editor, decimal, date, bool и т.д. Inline - это тип, который хранится в строке продукта (для ускорения выборки).
}

// Создать или удалить inline атрибут - нельзя

// Создание нового атрибута
func (ea *EavAttribute) create() error {

	// 1. Провекра на уникальность
	if !db.Unscoped().First(&EavAttribute{},"account_id = ? AND code = ?", ea.AccountID, ea.Code).RecordNotFound() {
		return utils.Error{Message: fmt.Sprintf("Атрибут с кодом = [%v] уже существует",ea.Code) }
	}


	return db.Create(ea).Error

	//return nil
}