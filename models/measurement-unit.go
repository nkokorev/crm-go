package models

import (
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)
// Единица измерения товара: штуки, метры, литры, граммы и т.д.
type MeasurementUnit struct {
	Id     		uint   `json:"id" gorm:"primaryKey"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`
	
	Name 		string `json:"name" gorm:"type:varchar(128);"` // штука, коробка, комплект, киллограмм, грамм,
	ShortName 	string `json:"short_name" gorm:"type:varchar(128);"` // шт., кор., компл., кг, гр,

	// ### Настройки ед. измерения 

	// весовой или нет
	Weighed 		bool `json:"weighed" gorm:"type:bool;default:false"`

	// тег для поиска
	Tag 		string `json:"tag" gorm:"type:varchar(32);"`

	// Описание (нужно ли?)
	Description *string `json:"description" gorm:"type:varchar(255);"`
}

func (MeasurementUnit) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&MeasurementUnit{}); err != nil {log.Fatal(err)}
	// db.Model(&MeasurementUnit{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE measurement_units ADD CONSTRAINT measurement_units_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

	mainAccount, err := GetMainAccount()
	if err != nil {
		log.Println("Не удалось найти главный аккаунт для MeasurementUnit")
	}

	measurementUnits := []MeasurementUnit{
		{Name:"штука", 			ShortName: "шт.", 		Weighed: false, 	Tag: "piece" },
		{Name:"коробка", 		ShortName: "кор.", 		Weighed: false, 	Tag: "box"},
		{Name:"упаковка", 		ShortName: "упак.", 	Weighed: false, 	Tag: "package"},
		{Name:"комплект", 		ShortName: "компл.", 	Weighed: false, 	Tag: "kit"},

		{Name:"грамм", 			ShortName: "гр.", 		Weighed: true, 	Tag: "gram"},
		{Name:"килограмм", 		ShortName: "кг.", 		Weighed: true, 	Tag: "kilogram"},
		{Name:"тонн", 			ShortName: "т.", 		Weighed: true, 	Tag: "tone"},

		{Name:"погонный метр", 	ShortName: "пог.м.", 	Weighed: false, 	Tag: "linearMeter"},
		{Name:"метр квадратный",ShortName: "м2.", 		Weighed: false, 	Tag: "squareMeter"},
		{Name:"метр кубический",ShortName: "м3.", 		Weighed: false, 	Tag: "cubicMeter"},

		{Name:"литр", 			ShortName: "л.", 		Weighed: false, 	Tag: "liter"},
		{Name:"миллилитр", 		ShortName: "мл.", 		Weighed: false, 	Tag: "milliliter"},
	}
	for _,v := range measurementUnits {
		_, err = mainAccount.CreateEntity(&v)
		if err != nil {
			log.Fatal(err)
		}
	}
}
func (measurementUnit *MeasurementUnit) BeforeCreate(tx *gorm.DB) error {
	measurementUnit.Id = 0
	return nil
}
func (measurementUnit *MeasurementUnit) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(measurementUnit)
	} else {
		_db = _db.Model(&MeasurementUnit{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{""})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}

// ############# Entity interface #############
func (measurementUnit MeasurementUnit) GetId() uint { return measurementUnit.Id }
func (measurementUnit *MeasurementUnit) setId(id uint) { measurementUnit.Id = id }
func (measurementUnit *MeasurementUnit) setPublicId(id uint) { measurementUnit.Id = id }
func (measurementUnit MeasurementUnit) GetAccountId() uint { return measurementUnit.AccountId }
func (measurementUnit *MeasurementUnit) setAccountId(id uint) { measurementUnit.AccountId = id }
func (measurementUnit MeasurementUnit) SystemEntity() bool { return measurementUnit.AccountId == 1 }

// ############# Entity interface #############


// ######### CRUD Functions ############
func (measurementUnit MeasurementUnit) create() (Entity, error)  {

	_item := measurementUnit
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}

func (MeasurementUnit) get(id uint, preloads []string) (Entity, error) {

	var item MeasurementUnit

	err := item.GetPreloadDb(false, false, preloads).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
func (measurementUnit *MeasurementUnit) load(preloads []string) error {
	if measurementUnit.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить MeasurementUnit - не указан  Id"}
	}

	err := measurementUnit.GetPreloadDb(false, false, preloads).First(measurementUnit, measurementUnit.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (*MeasurementUnit) loadByPublicId(preloads []string) error {
	return utils.Error{Message: "Нельзя загрузить по public Id"}
}
func (MeasurementUnit) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return MeasurementUnit{}.getPaginationList(accountId, 0, 50, sortBy, "",nil, preload)
}
func (MeasurementUnit) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	measurementUnits := make([]MeasurementUnit,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&MeasurementUnit{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id IN (?)", []uint{1, accountId}).
			Find(&measurementUnits, "name ILIKE ? OR ShortName ILIKE ? OR description ILIKE ?", search,search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&MeasurementUnit{}).GetPreloadDb(false, false, nil).
			Where("account_id = ? AND name ILIKE ? OR ShortName ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err :=(&MeasurementUnit{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id IN (?)", []uint{1, accountId}).
			Find(&measurementUnits).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&MeasurementUnit{}).GetPreloadDb(false, false, nil).Where("account_id IN (?)", []uint{1, accountId}).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(measurementUnits))
	for i := range measurementUnits {
		entities[i] = &measurementUnits[i]
	}

	return entities, total, nil
}

func (measurementUnit *MeasurementUnit) update(input map[string]interface{}, preloads []string) error {

	// delete(input,"order")
	utils.FixInputHiddenVars(&input)
	/*if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id"}); err != nil {
		return err
	}*/

	if err := measurementUnit.GetPreloadDb(false, false, nil).Where("id = ?", measurementUnit.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	err := measurementUnit.GetPreloadDb(false,false, preloads).First(measurementUnit, measurementUnit.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (measurementUnit *MeasurementUnit) delete () error {
	return measurementUnit.GetPreloadDb(true,false,nil).Where("id = ?", measurementUnit.Id).Delete(measurementUnit).Error
}


