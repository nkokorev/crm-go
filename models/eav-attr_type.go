package models

import (
	"fmt"
	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"reflect"
)

// EAV-Атрибуты, предусмотренные в аккаунте: Размер одежды, Тип упаковки, Состав, Цвет и т.д.
type EavAttrType struct {
	ID     		uint   `json:"id" gorm:"primary_key"`
	AccountID 	uint `json:"-" gorm:"type:int;index;not null;"`

	Code 		string	`json:"code" gorm:"type:varchar(50);index;not null;"` // color, size, etc.
	Name 		string 	`json:"name" gorm:"type:varchar(50);default:null;"` // Цвет, килограмм
	ShortName 	string 	`json:"shortName" gorm:"type:varchar(50);default:null;"` // Цвет, кг.,
	Description string 	`json:"description" gorm:"type:varchar(255);default:null;"` // Описание параметра (может быть нужно для отображения)

	// Multiple 	bool 	`json:"multiple"`
	// Required 	bool 	`json:"required"`  // обязате

	// это данные о хранении и отображении атрибута
	AttrTypeCode string 	`json:"attr_type_code"` // varchar, text, int, decimal, bool, date
	// Products 	[]Product	`json:"products" gorm:"many2many:product_eav_attributes"`

	ValueID   	uint	`json:"-" gorm:"type:int;"`
	ValueTable 	string 	`json:"-" gorm:"type:varchar(50);index;not null;"` // eav_attribute_varchar, text, int, decimal, bool, date
}

func (EavAttrType) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&EavAttrType{})
	db.Model(&EavAttrType{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}

func (EavAttrType) TableName() string {
	return "eav_attr_types"
}

func (eatType *EavAttrType) BeforeCreate(scope *gorm.Scope) error {
	eatType.ID = 0
	return nil
}

// ######### CRUD Functions ############
func (input EavAttrType) create() (*EavAttrType, error)  {
	var eatType = input
	eatType.ValueTable = "eav_attr_" +  eatType.AttrTypeCode
	err := db.Create(&eatType).Error
	return &eatType, err
}

func (EavAttrType) get(id uint) (*EavAttrType, error) {

	eatType := EavAttrType{}

	if err := db.Model(&eatType).First(&eatType, id).Error; err != nil {
		return nil, err
	}

	return &eatType, nil
}

func (EavAttrType) getByCode(code string) (*EavAttrType, error) {

	eatType := EavAttrType{}

	if err := db.Model(&eatType).First(&eatType, "code = ?", code).Error; err != nil {
		return nil, err
	}

	return &eatType, nil
}

func (EavAttrType) getList(accountId uint) ([]EavAttrType, error) {

	eatTypes := make([]EavAttrType,0)

	err := db.Model(&EavAttrType{}).Find(&eatTypes, "account_id = ?", accountId).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return eatTypes, nil
}

func (eatType *EavAttrType) update(input interface{}) error {
	return db.Model(eatType).Omit("id", "account_id").Updates(structs.Map(input)).Error
}

func (eatType EavAttrType) delete () error {
	return db.Model(EavAttrType{}).Where("id = ?", eatType.ID).Delete(eatType).Error
}
// ######### END CRUD Functions ############

// ######### ACCOUNT Functions ############
func (account Account) CreateEavAttrType(input EavAttrType) (*EavAttrType, error) {
	input.AccountID = account.ID

	if input.ExistCode() {
		return nil, utils.Error{Message: "Повторение данных", Errors: map[string]interface{}{"code":"Атрибут с таким code уже есть"}}
	}

	eatType, err := input.create()
	if err != nil {
		return nil, err
	}

	return eatType, nil
}

func (account Account) GetEavAttrType(eatId uint) (*EavAttrType, error) {
	eatType, err := EavAttrType{}.get(eatId)
	if err != nil {
		return nil, err
	}

	if account.ID != eatType.AccountID {
		return nil, utils.Error{Message: "Атрибут принадлежит другому аккаунту"}
	}

	return eatType, nil
}

func (account Account) GetEavAttrTypeByCode(code string) (*EavAttrType, error) {
	eatType, err := EavAttrType{}.getByCode(code)
	if err != nil {
		return nil, err
	}

	if account.ID != eatType.AccountID {
		return nil, utils.Error{Message: "Атрибут принадлежит другому аккаунту"}
	}

	return eatType, nil
}

func (account Account) GetEavAttrTypes() ([]EavAttrType, error) {

	eatTypes := make([]EavAttrType,0)

	err := db.Model(&EavAttrType{}).
		// Preload("ProductCards").
		// Joins("LEFT JOIN users ON account_users.user_id = users.id").
		Find(&eatTypes, "account_id = ?", account.ID).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, err
	}

	return eatTypes, nil
}

func (account Account) UpdateEavAttrType(eatTypeId uint, input interface{}) (*EavAttrType, error) {
	eatType, err := account.GetEavAttrType(eatTypeId)
	if err != nil {
		return nil, err
	}

	if account.ID != eatType.AccountID {
		return nil, utils.Error{Message: "Атрибут принадлежит другому аккаунту"}
	}

	err = eatType.update(input)

	return eatType, err

}

func (account Account) DeleteEavAttrType(eatTypeId uint) error {

	// включает в себя проверку принадлежности к аккаунту
	eatType, err := account.GetEavAttrType(eatTypeId)
	if err != nil {
		return err
	}

	return eatType.delete()
}

func (account Account) CreateBaseEavAttrTypes() error {

	// 2. // varchar, text, int, decimal, bool, date
	attrs := []EavAttrType{
		{Code: 	"color",		Name:	"Цвет", AttrTypeCode: "varchar"}, // белый, черный, венге и т.д.
		{Code:	"bodyMaterial", Name: "Материал корпуса", 	AttrTypeCode: "varchar"}, // металл
		{Code:	"filterType",  	Name: "Тип фильтра", 		AttrTypeCode: "varchar"}, // угольно-фотокаталитический
		{Code:	"performance",  Name: "Производительность", AttrTypeCode: "int", ShortName: "м³/ч"}, // 150 (м³/ч)
		{Code:	"rangeUVRadiation", Name: "Диапазон бактерицидного УФ излучения",  	AttrTypeCode: "varchar"}, // 250-260Нм
		{Code:	"powerLamp",  		Name: "Мощность излучения лампы рециркулятора",	AttrTypeCode: "varchar", ShortName: "Вт/м²"}, // 10,8 Вт/м²
		{Code:	"powerConsumption", Name: "Потребляемая мощность", 	AttrTypeCode: "int", ShortName: "Вт"}, // 60 (Вт)
		{Code:	"lifeTimeDevice",  	Name: "Срок службы устройства", AttrTypeCode: "int", ShortName: "ч"}, // 100000 (ч)
		{Code:	"lifeTimeLamp",  	Name: "Срок службы УФ лампы", 	AttrTypeCode: "int", ShortName: "ч"}, // 9000 (ч)
		{Code:	"baseTypeLamp",  	Name: "Тип цоколя лампы", 	AttrTypeCode: "varchar"}, // G13
		{Code:	"degreeProtection", Name: "Степень защиты", 	AttrTypeCode: "varchar"}, // IP20
		{Code:	"supplyVoltage",  	Name: "Напряжение питания", AttrTypeCode: "varchar"}, // 175-265В
		{Code:	"temperatureMode",  Name: "Температурный режим работы",AttrTypeCode: "varchar"}, // +2 +50С
		{Code:	"overallDimensions",Name: "Габаритные размеры(ВхШхГ)", AttrTypeCode: "varchar"}, //  690х250х250мм
		{Code:	"noiseLevel",  		Name: "Уровень шума", AttrTypeCode: "int", 	ShortName: "дБ"}, //  35 (дБ)
		{Code:	"grossWeight",  	Name: "Вес брутто", AttrTypeCode: "decimal",ShortName: "кг"}, // 5.5 (кг)
	}

	for i, _ := range attrs {
		_, err := account.CreateEavAttrType(attrs[i])
		if err != nil {
			fmt.Println("Cannot create EavAttrType: ", err)
			return err
		}
	}

	return nil
}
// ######### END ACCOUNT Functions ############

// ########## SELF FUNCTIONAL ############
func (eatType EavAttrType) ExistCode() bool {
	return !db.Unscoped().First(&EavAttrType{},"account_id = ? AND code = ?", eatType.AccountID, eatType.Code).RecordNotFound()
}

// func (product Product) CreateAttrValue(eatType EavAttrType, value Eav) (*EavAttribute, error) {
func (product Product) CreateAttrValue(eat EavAttribute, value Eav) (*EavAttribute, error) {

	// var eat EavAttribute // return value

	/*switch eatType.AttrTypeCode {
	case "varchar":
		var input EavAttrVarchar

		// проверяем переданный тип
		if reflect.TypeOf(val).Name() != "string" {
			return nil, utils.Error{Message: "Не верно указан тип string"}
		}
		input.Value = value.(string)

		// создаем данные в соотв. таблице
		_v, err := input.create()
		if err != nil {
			return nil, utils.Error{Message: "ошибка создания атрибута"}
		}

		valI := Eav(&_v)

		break
	case "int":
		var input EavAttrInt

		// проверяем переданный тип
		if reflect.TypeOf(val).Name() != "int64" {
			return nil, utils.Error{Message: "Не верно указан тип string"}
		}
		input.Value = value.(int64)

		// создаем данные в соотв. таблице
		val, err := input.create()
		if err != nil {
			return nil, utils.Error{Message: "ошибка создания атрибута"}
		}

		break
	case "decimal":
		var input EavAttrDecimal

		// проверяем переданный тип
		if reflect.TypeOf(val).Name() != "int64" {
			return nil, utils.Error{Message: "Не верно указан тип string"}
		}
		input.Value = value.(float64)

		// создаем данные в соотв. таблице
		val, err := input.create()
		if err != nil {
			return nil, utils.Error{Message: "ошибка создания атрибута"}
		}

		break
	}

	eat.ProductId = product.ID
	eat.EavAttrTypeId = val.getId()*/

	return eat.create()
	
}
