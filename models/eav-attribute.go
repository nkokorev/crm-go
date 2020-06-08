package models

import (
	"fmt"
	"github.com/nkokorev/crm-go/utils"
)

// EAV-Атрибуты, предусмотренные в аккаунте: Размер одежды, Тип упаковки, Состав, Цвет и т.д.
type EavAttribute struct {
	ID     		uint   	`json:"id" gorm:"primary_key"`
	AccountID 	uint 	`json:"-" gorm:"type:int;index;not_null;"`

	Code 		string	`json:"code"` // color
	Name 		string 	`json:"name"` // Цвет, килограмм
	ShortName 	string 	`json:"shortName"` // Цвет, кг.,
	Description string 	`json:"description" gorm:"type:varchar(255);"` // Описание параметра (может быть нужно для отображения)

	// Multiple 	bool 	`json:"multiple"`
	// Required 	bool 	`json:"required"`  // обязате

	// это данные о хранении и отображении атрибута
	AttrTypeCode string `json:"attr_type_code"` // varchar, text, int, decimal, bool, date | text_field, text_editor, decimal, date, bool и т.д.
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