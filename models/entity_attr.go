package models

import (
	"fmt"
	"github.com/nkokorev/crm-go/database/base"
)

// Таблица собирающая данные об атрибуте
type EavAttr struct {
	ID        			uint 				`gorm:"primary_key;unique_index;" json:"-"`
	HashID 				string 				`json:"hash_id" gorm:"type:varchar(10);unique_index;not null;"`
	AccountID 			uint 				`json:"-" gorm:"default:null;"` // если роль системная, то нет accountID
	Code 				string 				`json:"code" gorm:"not null;"` // "raw_material", "color", "size", "weight"
	Label 				string 				`json:"label" gorm:"not null;"` // "Тип сырья", "Цвет", "Размер", "Вес"
	Required 			bool 				`json:"required" gorm:"default:false;not null"` // обязателен ли к заполнению
	System 				bool 				`json:"system" gorm:"default:false;not null"` // дефолтный атрибут или нет
	Visible 			bool 				`json:"visible" gorm:"default:true;not null"` // отображать ли атрибут в карточке или еще где-либо
	Searchable 			bool 				`json:"visible" gorm:"default:true;not null;description:'Описание... some text there'"` // осуществлять ли поиск по данному атрибуту (в фильтрах или еще где-то)
	EavAttrInputTypeID	uint				`json:"eav_attr_input_type_id"`
	EavAttrType 		EavAttrType 		`json:"eav_attr_type" gorm:""` // дефолтный атрибут или нет (дефолтный нельзя удалить)
	EavAttrTypeID 		uint 				`json:"-" gorm:""` // дефолтный атрибут или нет (дефолтный нельзя удалить)
	EavAttrSets 		[]EavAttrSet 		`json:"eav_attr_groups" gorm:"many2many:eav_attr_eav_attr_sets;"`
	Products 			[]Product 			`json:"eav_attr_products" gorm:"many2many:eav_attr_products;"` // очень важная фича для поиска
}

// Типы атрибутов
type EavAttrType struct {
	ID        	uint 	`gorm:"primary_key;unique_index;" json:"-"`
	HashID 		string 	`json:"hash_id" gorm:"type:varchar(10);unique_index;not null;"`
	EavAttrs		[]EavAttr	`json:"eav_attrs"` // обратная связь
}

// Типы атрибутов
type EavAttrInputType struct {
	ID        	uint 		`gorm:"primary_key;unique_index;" json:"-"`
	Property 	string 		`json:"property"` // Text Field, Text Area, Text Editor, Date, Yes/No (Boolean), Dropdown, Multiple Select, Price, Media Image
}

//
type EavAttrVarchar struct {
	ID        	uint 	`gorm:"primary_key;unique_index;" json:"-"`
	HashID 		string 	`json:"hash_id" gorm:"type:varchar(10);unique_index;not null;"`
	EavAttrID 	uint 	`json:"-" gorm:"not null;"`
	Value 		string 	`json:"name" gorm:"not null;"` // "Тип сырья", "Цвет", "Размер", "Вес"
}

/*type EavAttrGroup struct {
	ID        	uint 	`gorm:"primary_key;unique_index;" json:"-"`
	HashID 		string 	`json:"hash_id" gorm:"type:varchar(10);unique_index;not null;"`
	Name 		string 	`json:"name" gorm:"not null;"` // "Тип сырья", "Цвет", "Размер", "Вес"
	//EavAttrs 	[]EavAttr 	`json:"eav_attrs"`
	EavAttrs 	[]EavAttr 	`json:"eav_attrs" gorm:"many2many:eav_attr_eav_attr_groups;"`
	Product		[]Product 	`gorm:"many2many:eav_attr_group_products;"` // продукты соотносятся с группами атрибутов
}*/

type EavAttrSet struct {
	ID        	uint 	`gorm:"primary_key;unique_index;" json:"-"`
	HashID 		string 	`json:"hash_id" gorm:"type:varchar(10);unique_index;not null;"`
	Name 		string 	`json:"name" gorm:"not null;"` // "Тип сырья", "Цвет", "Размер", "Вес"
	//EavAttrs 	[]EavAttr 	`json:"eav_attrs"`
	EavAttrs 	[]EavAttr 	`json:"eav_attrs" gorm:"many2many:eav_attr_eav_attr_sets;"`
	Products	[]Product 	`json:"products" gorm:"many2many:eav_attr_set_products;"` // продукты соотносятся с группами атрибутов
}

/*
func (EavAttr) TableName() string {
	return "eav_attr"
}
*/

func (EavAttrVarchar) TableName() string {
	return "eav_attr_varchar"
}
/*func (EavAttrSet) TableName() string {
	return "eav_attr_set"
}*/

func (e *EavAttr) CreateHashID() error {
	return nil
}

func CreateSystemEavAttr() {
	db := base.GetDB()
	db.Unscoped().Delete(&Role{})

	eav_attrs := []EavAttr{
		{Code: "color", Label: "Цвет", System:	true, Searchable: true}, // one of
		{Code: "date_of_manufacture", Label: "Дата изготовления", System:	true, Searchable: true}, // datetime
		{Code: "country_of_manufacture", Label: "Страна изготовления", System:	true, Searchable: true}, // on of Country
	}

	// создаем системные роли
	for _, v := range eav_attrs {

		fmt.Println(v)
	}
}
