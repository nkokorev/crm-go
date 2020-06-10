package models

import (
	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
)

// EAV-Атрибуты, предусмотренные в аккаунте: Размер одежды, Тип упаковки, Состав, Цвет и т.д.
type EavAttrInt struct {
	ID     			uint   	`json:"id" gorm:"primary_key"`
	EavAttributeId	uint	`json:"eavAttributeId" gorm:"type:int;index;not null;"` // forkey
	Value 			int64		`json:"value" gorm:"type:int;default:null;"` // color, size, etc.
}

func (EavAttrInt) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&EavAttrInt{})
	db.Model(&EavAttrInt{}).AddForeignKey("eav_attribute_id", "eav_attributes(id)", "CASCADE", "CASCADE")

}

func (EavAttrInt) TableName() string {
	return "eav_attr_int"
}

func (eatInt *EavAttrInt) BeforeCreate(scope *gorm.Scope) error {
	eatInt.ID = 0
	return nil
}

// ######### CRUD Functions ############
func (input EavAttrInt) create() (*EavAttrInt, error)  {
	var eatInt = input
	err := db.Create(&eatInt).Error
	return &eatInt, err
}

func (EavAttrInt) get(id uint) (*EavAttrInt, error) {

	eatInt := EavAttrInt{}

	if err := db.Model(&eatInt).First(&eatInt, id).Error; err != nil {
		return nil, err
	}

	return &eatInt, nil
}

func (eatInt *EavAttrInt) update(input interface{}) error {
	return db.Model(eatInt).Omit("id", "account_id").Updates(structs.Map(input)).Error
}

func (eatInt EavAttrInt) delete () error {
	return db.Model(EavAttrInt{}).Where("id = ?", eatInt.ID).Delete(eatInt).Error
}

// ######### END CRUD Functions ############

func (eatInt EavAttrInt) getId() uint {
	return eatInt.ID
}

// ######### END CRUD Functions ############
