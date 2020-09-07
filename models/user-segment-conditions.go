package models

import (
	"database/sql"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

type UserSegmentCondition struct {
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	PublicId	uint   	`json:"public_id" gorm:"type:int;index;not null;default:1"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	// Имя кампании сегмента: 'активные участники', 'только клиенты', 'майские подписчики'
	Name 		string	`json:"name" gorm:"type:varchar(128);"`

	// Объекты для определения типа соответствия
	OwnerId		uint	`json:"owner_id" gorm:"type:int;not null;"` // Id в
	OwnerType	string 	`json:"owner_type" gorm:"type:varchar(255);not null;"`

	// =============   Настройки получателей    ===================
	CreatedAt	time.Time `json:"created_at"`
}

// ############# Entity interface #############
func (userSegmentCondition UserSegmentCondition) GetId() uint { return userSegmentCondition.Id }
func (userSegmentCondition *UserSegmentCondition) setId(id uint) { userSegmentCondition.Id = id }
func (userSegmentCondition *UserSegmentCondition) setPublicId(publicId uint) { userSegmentCondition.PublicId = publicId }
func (userSegmentCondition UserSegmentCondition) GetAccountId() uint { return userSegmentCondition.AccountId }
func (userSegmentCondition *UserSegmentCondition) setAccountId(id uint) { userSegmentCondition.AccountId = id }
func (UserSegmentCondition) SystemEntity() bool { return false }

// ############# Entity interface #############

func (UserSegmentCondition) PgSqlCreate() {
	if err := db.Migrator().AutoMigrate(&UserSegmentCondition{}); err != nil { log.Fatal(err)}
	// db.Model(&UserSegmentCondition{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE user_segment_conditions ADD CONSTRAINT users_segment_conditions_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (userSegmentCondition *UserSegmentCondition) BeforeCreate(tx *gorm.DB) error {
	userSegmentCondition.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&UserSegmentCondition{}).Where("account_id = ?",  userSegmentCondition.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	userSegmentCondition.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (userSegmentCondition *UserSegmentCondition) AfterFind(tx *gorm.DB) (err error) {
	return nil
}

// ######### CRUD Functions ############
func (userSegmentCondition UserSegmentCondition) create() (Entity, error)  {

	en := userSegmentCondition

	if err := db.Create(&en).Error; err != nil {
		return nil, err
	}

	err := en.GetPreloadDb(false,false, nil).First(&en, en.Id).Error
	if err != nil {
		return nil, err
	}

	var newItem Entity = &en

	return newItem, nil
}

func (UserSegmentCondition) get(id uint, preloads []string) (Entity, error) {

	var userSegmentCondition UserSegmentCondition

	err := userSegmentCondition.GetPreloadDb(false,false,preloads).First(&userSegmentCondition, id).Error
	if err != nil {
		return nil, err
	}
	return &userSegmentCondition, nil
}
func (userSegmentCondition *UserSegmentCondition) load(preloads []string) error {

	err := userSegmentCondition.GetPreloadDb(false,false,preloads).First(userSegmentCondition, userSegmentCondition.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (userSegmentCondition *UserSegmentCondition) loadByPublicId(preloads []string) error {

	if userSegmentCondition.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить UserSegmentCondition - не указан  Id"}
	}
	if err := userSegmentCondition.GetPreloadDb(false,false, preloads).First(userSegmentCondition, "account_id = ? AND public_id = ?", userSegmentCondition.AccountId, userSegmentCondition.PublicId).Error; err != nil {
		return err
	}

	return nil
}

func (UserSegmentCondition) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return UserSegmentCondition{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (UserSegmentCondition) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	userSegmentCondition := make([]UserSegmentCondition,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&UserSegmentCondition{}).GetPreloadDb(false,false,preloads).
			Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&userSegmentCondition, "name ILIKE ? OR subject ILIKE ? OR preview_text ILIKE ?", search,search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&UserSegmentCondition{}).
			Where("account_id = ? AND name ILIKE ? OR description ILIKE ? ", accountId, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&UserSegmentCondition{}).GetPreloadDb(false,false,preloads).
			Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&userSegmentCondition).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&UserSegmentCondition{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(userSegmentCondition))
	for i := range userSegmentCondition {
		entities[i] = &userSegmentCondition[i]
	}

	return entities, total, nil
}

func (userSegmentCondition *UserSegmentCondition) update(input map[string]interface{}, preloads []string) error {

	if err := userSegmentCondition.GetPreloadDb(true,false,nil).Where(" id = ?", userSegmentCondition.Id).
		Omit("id", "account_id","created_at").Updates(input).Error; err != nil {
		return err
	}

	err := userSegmentCondition.GetPreloadDb(true,false,preloads).First(userSegmentCondition, userSegmentCondition.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (userSegmentCondition *UserSegmentCondition) delete () error {
	return userSegmentCondition.GetPreloadDb(true,false,nil).Where("id = ?", userSegmentCondition.Id).Delete(userSegmentCondition).Error
}
// ######### END CRUD Functions ############
func (userSegmentCondition *UserSegmentCondition) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(&userSegmentCondition)
	} else {
		_db = _db.Model(&UserSegmentCondition{})
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

// Получения списка пользователей в сегменте
func (userSegmentCondition UserSegmentCondition) Execute(data map[string]interface{}) error {
	

	return nil
}
