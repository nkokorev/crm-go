package models

import (
	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
)

// EAV-Атрибуты, предусмотренные в аккаунте: Размер одежды, Тип упаковки, Состав, Цвет и т.д.
type EavAttrVarchar struct {
	ID     			uint   `json:"id" gorm:"primary_key"`
	EavAttrId	uint	`json:"eavAttrId" gorm:"type:int;index;not null;"` // forkey
	Value 			string	`json:"value" gorm:"type:varchar(255);default:null;"` // color, size, etc.
}

func (EavAttrVarchar) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&EavAttrVarchar{})
	db.Model(&EavAttrVarchar{}).AddForeignKey("eav_attribute_id", "eav_attributes(id)", "CASCADE", "CASCADE")

}

func (EavAttrVarchar) TableName() string {
	return "eav_attr_varchar"
}

func (eatVarchar *EavAttrVarchar) BeforeCreate(scope *gorm.Scope) error {
	eatVarchar.ID = 0
	return nil
}

// ######### CRUD Functions ############
func (input EavAttrVarchar) create() (interface{}, error)  {
	var eatVarchar = input
	err := db.Create(&eatVarchar).Error
	return &eatVarchar, err
}

func (EavAttrVarchar) get(id uint) (*EavAttrVarchar, error) {

	eatVarchar := EavAttrVarchar{}

	if err := db.Model(&eatVarchar).First(&eatVarchar, id).Error; err != nil {
		return nil, err
	}

	return &eatVarchar, nil
}

func (eatVarchar *EavAttrVarchar) update(input interface{}) error {
	return db.Model(eatVarchar).Omit("id", "account_id").Updates(structs.Map(input)).Error
}

func (eatVarchar EavAttrVarchar) delete () error {
	return db.Model(EavAttrVarchar{}).Where("id = ?", eatVarchar.ID).Delete(eatVarchar).Error
}

// ######### END CRUD Functions ############

func (eatVarchar EavAttrVarchar) getId() uint {
	return eatVarchar.ID
}

// ######### END CRUD Functions ############
