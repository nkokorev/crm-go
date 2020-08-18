package models

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type UserSegmentConditions struct {
	Id     			uint   	`json:"id" gorm:"primary_key"`
	PublicId		uint   	`json:"publicId" gorm:"type:int;index;not null;default:1"`
	AccountId 		uint 	`json:"-" gorm:"type:int;index;not null;"`

	// Имя кампании сегмента: 'активные участники', 'только клиенты', 'майские подписчики'
	Name 			string 		`json:"name" gorm:"type:varchar(128);"`

	// Объекты для определения типа соответствия
	OwnerId	uint	`json:"ownerId" gorm:"type:int;not null;"` // Id в
	OwnerType	string `json:"ownerType" gorm:"type:varchar(255);not null;"`

	// =============   Настройки получателей    ===================
	CreatedAt 		time.Time `json:"createdAt"`
}

// ############# Entity interface #############
func (userSegmentConditions UserSegmentConditions) GetId() uint { return userSegmentConditions.Id }
func (userSegmentConditions *UserSegmentConditions) setId(id uint) { userSegmentConditions.Id = id }
func (userSegmentConditions *UserSegmentConditions) setPublicId(publicId uint) { userSegmentConditions.PublicId = publicId }
func (userSegmentConditions UserSegmentConditions) GetAccountId() uint { return userSegmentConditions.AccountId }
func (userSegmentConditions *UserSegmentConditions) setAccountId(id uint) { userSegmentConditions.AccountId = id }
func (UserSegmentConditions) SystemEntity() bool { return false }

// ############# Entity interface #############

func (UserSegmentConditions) PgSqlCreate() {
	db.CreateTable(&UserSegmentConditions{})
	db.Model(&UserSegmentConditions{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&UserSegmentConditions{}).AddForeignKey("email_template_id", "email_templates(id)", "RESTRICT", "CASCADE")
}
func (userSegmentConditions *UserSegmentConditions) BeforeCreate(scope *gorm.Scope) error {
	userSegmentConditions.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&UserSegmentConditions{}).Where("account_id = ?",  userSegmentConditions.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	userSegmentConditions.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (userSegmentConditions *UserSegmentConditions) AfterFind() (err error) {
	return nil
}

// ######### CRUD Functions ############
func (userSegmentConditions UserSegmentConditions) create() (Entity, error)  {

	en := userSegmentConditions

	if err := db.Create(&en).Error; err != nil {
		return nil, err
	}

	err := en.GetPreloadDb(false,false, true).First(&en, en.Id).Error
	if err != nil {
		return nil, err
	}

	var newItem Entity = &en

	return newItem, nil
}

func (UserSegmentConditions) get(id uint) (Entity, error) {

	var userSegmentConditions UserSegmentConditions

	err := userSegmentConditions.GetPreloadDb(false,false,true).First(&userSegmentConditions, id).Error
	if err != nil {
		return nil, err
	}
	return &userSegmentConditions, nil
}
func (userSegmentConditions *UserSegmentConditions) load() error {

	err := userSegmentConditions.GetPreloadDb(false,false,true).First(userSegmentConditions, userSegmentConditions.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (userSegmentConditions *UserSegmentConditions) loadByPublicId() error {

	if userSegmentConditions.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить UserSegmentConditions - не указан  Id"}
	}
	if err := userSegmentConditions.GetPreloadDb(false,false, true).First(userSegmentConditions, "account_id = ? AND public_id = ?", userSegmentConditions.AccountId, userSegmentConditions.PublicId).Error; err != nil {
		return err
	}

	return nil
}

func (UserSegmentConditions) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return UserSegmentConditions{}.getPaginationList(accountId, 0, 100, sortBy, "",nil)
}
func (UserSegmentConditions) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{}) ([]Entity, uint, error) {

	userSegmentConditions := make([]UserSegmentConditions,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&UserSegmentConditions{}).GetPreloadDb(true,false,true).Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Preload("EmailTemplate", func(db *gorm.DB) *gorm.DB {
				return db.Select(EmailTemplate{}.SelectArrayWithoutData())
			}).Preload("EmailBox").
			Find(&userSegmentConditions, "name ILIKE ? OR subject ILIKE ? OR preview_text ILIKE ?", search,search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&UserSegmentConditions{}).
			Where("account_id = ? AND name ILIKE ? OR description ILIKE ? ", accountId, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&UserSegmentConditions{}).GetPreloadDb(true,false,true).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Preload("EmailTemplate", func(db *gorm.DB) *gorm.DB {
				return db.Select(EmailTemplate{}.SelectArrayWithoutData())
			}).Preload("EmailBox").
			Find(&userSegmentConditions).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&UserSegmentConditions{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(userSegmentConditions))
	for i := range userSegmentConditions {
		entities[i] = &userSegmentConditions[i]
	}

	return entities, total, nil
}

func (userSegmentConditions *UserSegmentConditions) update(input map[string]interface{}) error {

	if err := userSegmentConditions.GetPreloadDb(true,false,false).Where(" id = ?", userSegmentConditions.Id).
		Omit("id", "account_id","created_at").Updates(input).Error; err != nil {
		return err
	}

	err := userSegmentConditions.GetPreloadDb(true,false,false).First(userSegmentConditions, userSegmentConditions.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (userSegmentConditions *UserSegmentConditions) delete () error {
	return userSegmentConditions.GetPreloadDb(true,true,false).Where("id = ?", userSegmentConditions.Id).Delete(userSegmentConditions).Error
}
// ######### END CRUD Functions ############

func (userSegmentConditions *UserSegmentConditions) GetPreloadDb(autoUpdateOff bool, getModel bool, preload bool) *gorm.DB {
	_db := db

	if autoUpdateOff {
		_db = _db.Set("gorm:association_autoupdate", false)
	}
	if getModel {
		_db = _db.Model(userSegmentConditions)
	} else {
		_db = _db.Model(&UserSegmentConditions{})
	}

	if preload {
		return _db.Preload("EmailTemplate", func(db *gorm.DB) *gorm.DB {
			return db.Select(EmailTemplate{}.SelectArrayWithoutData())
		}).Preload("EmailBox")
	} else {
		return _db
	}
}

// Получения списка пользователей в сегменте
func (userSegmentConditions UserSegmentConditions) Execute(data map[string]interface{}) error {
	

	return nil
}
