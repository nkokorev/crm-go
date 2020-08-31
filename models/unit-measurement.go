package models

import (
	"fmt"
	"gorm.io/gorm"
	"log"
)

type UnitMeasurement struct {
	Id     uint   `json:"id" gorm:"primaryKey"`

	Name 		string `json:"name" gorm:"type:varchar(128);"` // штука, коробка, комплект, киллограмм, грамм,
	ShortName 	string `json:"short_name" gorm:"type:varchar(128);"` // шт., кор., компл., кг, гр,
	Weight 		bool // весовой или нет

	Tag 		string `json:"tag" gorm:"type:varchar(32);"` // для поиска

	Description *string `json:"description" gorm:"type:varchar(255);"` // pgsql: text
}

func (UnitMeasurement) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&UnitMeasurement{}); err != nil {
		log.Fatal(err)
	}

	// 2.
	units := []UnitMeasurement{
		{Name:"штука", ShortName: "шт.", Weight: false, Tag: "piece" },
		{Name:"коробка", ShortName: "кор.", Weight: false, Tag: "box"},
		{Name:"упаковка", ShortName: "упак.", Weight: false, Tag: "package"},
		{Name:"комплект", ShortName: "компл.", Weight: false, Tag: "kit"},
		{Name:"килограмм", ShortName: "кг.", Weight: true, Tag: "kilogram"},
		{Name:"грамм", ShortName: "гр.", Weight: true, Tag: "gram"},
		{Name:"погонный метр", ShortName: "пог.м.", Weight: false, Tag: "linearMeter"},
		{Name:"метр квадратный", ShortName: "м2.", Weight: false, Tag: "squareMeter"},
		{Name:"литр", ShortName: "л.", Weight: false, Tag: "liter"},
		{Name:"миллилитр", ShortName: "мл.", Weight: false, Tag: "milliliter"},
	}

	for i, _ := range units {
		_, err := units[i].create()
		if err != nil {
			fmt.Println("Cannot create UnitMeasurement: ", err)
		}
	}
}

func (um *UnitMeasurement) BeforeCreate(tx *gorm.DB) error {
	um.Id = 0
	return nil
}

func (UnitMeasurement) TableName() string {
	return "unit_measurements"
}

// ######### CRUD Functions ############
func (um UnitMeasurement) create() (*UnitMeasurement, error)  {
	var unit = um
	err := db.Create(&unit).Error
	return &unit, err
}

func (UnitMeasurement) get(id uint) (*UnitMeasurement, error) {

	unit := UnitMeasurement{}

	if err := db.First(&unit, id).Error; err != nil {
		return nil, err
	}

	return &unit, nil
}

func (UnitMeasurement) getList() ([]UnitMeasurement, error) {

	units := make([]UnitMeasurement,0)

	err := db.Find(&units).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return units, nil
}

func (ut *UnitMeasurement) update(input map[string]interface{}) error {
	return db.Model(ut).Omit("id").Updates(input).Error

}

func (ut *UnitMeasurement) delete () error {
	return db.Model(UnitMeasurement{}).Where("id = ?", ut.Id).Delete(ut).Error
}
// ######### END CRUD Functions ############
