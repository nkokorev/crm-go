package models

import (
	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
)

// EAV-Атрибуты, предусмотренные в аккаунте: Размер одежды, Тип упаковки, Состав, Цвет и т.д.
type EavAttrDecimal struct {
	ID     			uint   	`json:"id" gorm:"primary_key"`
	EavAttributeId	uint	`json:"eavAttributeId" gorm:"type:int;index;not null;"` // forkey
	Value 			float64		`json:"value" gorm:"type:decimal;default:null;"` // 12, 15.3
}

func (EavAttrDecimal) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&EavAttrDecimal{})
	db.Model(&EavAttrDecimal{}).AddForeignKey("eav_attribute_id", "eav_attributes(id)", "CASCADE", "CASCADE")

}

func (EavAttrDecimal) TableName() string {
	return "eav_attr_int"
}

func (eatInt *EavAttrDecimal) BeforeCreate(scope *gorm.Scope) error {
	eatInt.ID = 0
	return nil
}

// ######### CRUD Functions ############
func (input EavAttrDecimal) create() (*EavAttrDecimal, error)  {
	var eatInt = input
	err := db.Create(&eatInt).Error
	return &eatInt, err
}

func (EavAttrDecimal) get(id uint) (*EavAttrDecimal, error) {

	eatType := EavAttrDecimal{}

	if err := db.Model(&eatType).First(&eatType, id).Error; err != nil {
		return nil, err
	}

	return &eatType, nil
}

func (eatType *EavAttrDecimal) update(input interface{}) error {
	return db.Model(eatType).Omit("id", "account_id").Updates(structs.Map(input)).Error
}

func (eatType EavAttrDecimal) delete () error {
	return db.Model(EavAttrDecimal{}).Where("id = ?", eatType.ID).Delete(eatType).Error
}

// ######### END CRUD Functions ############

func (eatDecimal EavAttrDecimal) getId() uint {
	return eatDecimal.ID
}