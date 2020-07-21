package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

type UnitMeasurement struct {
	ID     uint   `json:"id" gorm:"primary_key"`

	Name 		string `json:"name" gorm:"type:varchar(128);"` // штука, коробка, комплект, киллограмм, грамм,
	ShotName 	string `json:"name" gorm:"type:varchar(128);"` // шт., кор., компл., кг, гр,
	Weight 		bool // весовой или нет

	Tag 		string `json:"tag" gorm:"type:varchar(32);"` // для поиска

	Description string `json:"description" gorm:"type:varchar(255);default:null;"` // pgsql: text
}

func (UnitMeasurement) PgSqlCreate() {
	db.CreateTable(&UnitMeasurement{})

	// 2.
	units := []UnitMeasurement{
		{Name:"штука", ShotName: "шт.", Weight: false, Tag: "piece" },
		{Name:"коробка", ShotName: "кор.", Weight: false, Tag: "box"},
		{Name:"упаковка", ShotName: "упак.", Weight: false, Tag: "package"},
		{Name:"комплект", ShotName: "компл.", Weight: false, Tag: "kit"},
		{Name:"килограмм", ShotName: "кг.", Weight: true, Tag: "kilogram"},
		{Name:"грамм", ShotName: "гр.", Weight: true, Tag: "gram"},
		{Name:"погонный метр", ShotName: "пог.м.", Weight: false, Tag: "linearMeter"},
		{Name:"метр квадратный", ShotName: "м2.", Weight: false, Tag: "squareMeter"},
		{Name:"литр", ShotName: "л.", Weight: false, Tag: "liter"},
		{Name:"миллилитр", ShotName: "мл.", Weight: false, Tag: "milliliter"},
	}

	for i, _ := range units {
		_, err := units[i].create()
		if err != nil {
			fmt.Println("Cannot create UnitMeasurement: ", err)
		}
	}
}

func (um *UnitMeasurement) BeforeCreate(scope *gorm.Scope) error {
	um.ID = 0
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

func (ut UnitMeasurement) delete () error {
	return db.Model(UnitMeasurement{}).Where("id = ?", ut.ID).Delete(ut).Error
}
// ######### END CRUD Functions ############
