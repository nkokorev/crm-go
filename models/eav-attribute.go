package models

import (
	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
)

// EAV-Атрибуты, предусмотренные в аккаунте: Размер одежды, Тип упаковки, Состав, Цвет и т.д.
type EavAttribute struct {
	ID     		uint   `json:"id" gorm:"primary_key"`

	ProductId 	uint `json:"productId" gorm:"type:int;index;not null;"`
	EavAttrTypeId uint `json:"eavAttrTypeId" gorm:"type:int;index;not null;"`
	ValueID   	uint	`json:"ownerId" gorm:"type:int;"` // id in attributes table
	
	Value interface{} `json:"value" sql:"-"`
}

func (EavAttribute) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&EavAttribute{})
	db.Model(&EavAttrType{}).AddForeignKey("product_id", "products(id)", "CASCADE", "CASCADE")
	db.Model(&EavAttrType{}).AddForeignKey("eav_attribute_type_id", "eav_attr_types(id)", "CASCADE", "CASCADE")

	// db.Exec("ALTER TABLE eav_attributes\n    ADD CONSTRAINT eav_attributes_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\ncreate unique index uix_eav_attributes_account_id_code ON eav_attributes (account_id,code);\n")

}

func (eat *EavAttribute) BeforeCreate(scope *gorm.Scope) error {
	eat.ID = 0
	return nil
}

// ######### CRUD Functions ############
func (input EavAttribute) create() (*EavAttribute, error)  {
	var eat = input
	// eat.OwnerType = "eav_attribute_" +  eat.AttrTypeCode
	err := db.Create(&eat).Error
	return &eat, err
}

func (EavAttribute) get(id uint) (*EavAttribute, error) {

	eat := EavAttribute{}

	if err := db.Model(&eat).First(&eat, id).Error; err != nil {
		return nil, err
	}

	return &eat, nil
}

func (EavAttribute) getByCode(code string) (*EavAttribute, error) {

	eat := EavAttribute{}

	if err := db.Model(&eat).First(&eat, "code = ?", code).Error; err != nil {
		return nil, err
	}

	return &eat, nil
}

func (EavAttribute) getList(accountId uint) ([]EavAttribute, error) {

	eats := make([]EavAttribute,0)

	err := db.Model(&EavAttribute{}).Find(&eats, "account_id = ?", accountId).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return eats, nil
}

func (eat *EavAttribute) update(input interface{}) error {
	return db.Model(eat).Omit("id", "account_id").Updates(structs.Map(input)).Error
}

func (eat EavAttribute) delete () error {
	return db.Model(EavAttribute{}).Where("id = ?", eat.ID).Delete(eat).Error
}
// ######### END CRUD Functions ############

// ######### ACCOUNT Functions ############


/*func (account Account) GetEavAttribute(eatId uint) (*EavAttribute, error) {
	eat, err := EavAttribute{}.get(eatId)
	if err != nil {
		return nil, err
	}

	if account.ID != eat.AccountID {
		return nil, utils.Error{Message: "Атрибут принадлежит другому аккаунту"}
	}

	return eat, nil
}

func (account Account) GetEavAttributeByCode(code string) (*EavAttribute, error) {
	eat, err := EavAttribute{}.getByCode(code)
	if err != nil {
		return nil, err
	}

	if account.ID != eat.AccountID {
		return nil, utils.Error{Message: "Атрибут принадлежит другому аккаунту"}
	}

	return eat, nil
}

func (account Account) GetEavAttributes() ([]EavAttribute, error) {

	eats := make([]EavAttribute,0)

	err := db.Model(&EavAttribute{}).
		// Preload("ProductCards").
		// Joins("LEFT JOIN users ON account_users.user_id = users.id").
		Find(&eats, "account_id = ?", account.ID).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, err
	}

	return eats, nil
}

func (account Account) UpdateEavAttribute(eatId uint, input interface{}) (*EavAttribute, error) {
	eat, err := account.GetEavAttribute(eatId)
	if err != nil {
		return nil, err
	}

	if account.ID != eat.AccountID {
		return nil, utils.Error{Message: "Атрибут принадлежит другому аккаунту"}
	}

	err = eat.update(input)

	return eat, err

}

func (account Account) DeleteEavAttribute(eatId uint) error {

	// включает в себя проверку принадлежности к аккаунту
	eat, err := account.GetEavAttribute(eatId)
	if err != nil {
		return err
	}

	return eat.delete()
}

func (account Account) CreateBaseEavAttributes() error {

	// 2. // varchar, text, int, decimal, bool, date
	attrs := []EavAttribute{
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
		_, err := account.CreateEavAttribute(attrs[i])
		if err != nil {
			fmt.Println("Cannot create EavAttribute: ", err)
			return err
		}
	}

	return nil
}*/
// ######### END ACCOUNT Functions ############

// ########## SELF FUNCTIONAL ############

type Attribute interface {
	create()
	getId()
	get(id uint)
	// getByProduct(productId uint)
}