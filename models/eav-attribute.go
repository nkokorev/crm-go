package models

import (
	"fmt"
	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
)

// EAV-Атрибуты, предусмотренные в аккаунте: Размер одежды, Тип упаковки, Состав, Цвет и т.д.
type EavAttribute struct {
	ID     		uint   `json:"id" gorm:"primary_key"`
	AccountID 	uint `json:"-" gorm:"type:int;index;not null;"`

	Code 		string	`json:"code" gorm:"type:varchar(50);index;not null;"` // color, size, etc.
	Name 		string 	`json:"name" gorm:"type:varchar(50);default:null;"` // Цвет, килограмм
	ShortName 	string 	`json:"shortName" gorm:"type:varchar(50);default:null;"` // Цвет, кг.,
	Description string 	`json:"description" gorm:"type:varchar(255);default:null;"` // Описание параметра (может быть нужно для отображения)

	// Multiple 	bool 	`json:"multiple"`
	// Required 	bool 	`json:"required"`  // обязате

	// это данные о хранении и отображении атрибута
	AttrTypeCode string `json:"attr_type_code"` // varchar, text, int, decimal, bool, date 
}

func (EavAttribute) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&EavAttribute{})
	db.Exec("ALTER TABLE eav_attributes\n    ADD CONSTRAINT eav_attributes_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\ncreate unique index uix_eav_attributes_account_id_code ON eav_attributes (account_id,code);\n")

}

func (eat *EavAttribute) BeforeCreate(scope *gorm.Scope) error {
	eat.ID = 0
	return nil
}

// ######### CRUD Functions ############
func (input EavAttribute) create() (*EavAttribute, error)  {
	var eat = input
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
func (account Account) CreateEavAttribute(input EavAttribute) (*EavAttribute, error) {
	input.AccountID = account.ID

	if input.ExistCode() {
		return nil, utils.Error{Message: "Повторение данных", Errors: map[string]interface{}{"code":"Атрибут с таким code уже есть"}}
	}
	
	eat, err := input.create()
	if err != nil {
		return nil, err
	}

	return eat, nil
}

func (account Account) GetEavAttribute(eatId uint) (*EavAttribute, error) {
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

	// 2.
	attrs := []EavAttribute{
		{Code: 	"color", Name:"Цвет", AttrTypeCode: "varchar"},
		{Name:	"size",  Code: "Размер", AttrTypeCode: "decimal"},
		{Name:	"bodyMaterial",  Code: "Материал корпуса", AttrTypeCode: "varchar"},
		{Name:	"filterType",  Code: "Тип фильтра", AttrTypeCode: "varchar"},
	}

	for i, _ := range attrs {
		_, err := account.CreateEavAttribute(attrs[i])
		if err != nil {
			fmt.Println("Cannot create EavAttribute: ", err)
			return err
		}
	}

	return nil
}
// ######### END ACCOUNT Functions ############

// ########## SELF FUNCTIONAL ############
func (eat EavAttribute) ExistCode() bool {
	return !db.Unscoped().First(&EavAttribute{},"account_id = ? AND code = ?", eat.AccountID, eat.Code).RecordNotFound()
}
