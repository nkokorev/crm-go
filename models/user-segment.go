package models

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type UserSegment struct {
	Id     			uint   	`json:"id" gorm:"primary_key"`
	PublicId		uint   	`json:"publicId" gorm:"type:int;index;not null;default:1"`
	AccountId 		uint 	`json:"-" gorm:"type:int;index;not null;"`

	// Имя кампании сегмента: 'активные участники', 'только клиенты', 'майские подписчики'
	Name 			string 		`json:"name" gorm:"type:varchar(128);"`

	// AND or ANY
	StrictMatching 		bool 	`json:"enabled" gorm:"type:bool;default:false;"`

	// персональные настройки сегмента
	UserSegmentRules []UserSegmentRule `json:"userSegmentRules"`

	// =============   Настройки получателей    ===================
	CreatedAt 		time.Time `json:"createdAt"`
}

// ############# Entity interface #############
func (userSegment UserSegment) GetId() uint { return userSegment.Id }
func (userSegment *UserSegment) setId(id uint) { userSegment.Id = id }
func (userSegment *UserSegment) setPublicId(publicId uint) { userSegment.PublicId = publicId }
func (userSegment UserSegment) GetAccountId() uint { return userSegment.AccountId }
func (userSegment *UserSegment) setAccountId(id uint) { userSegment.AccountId = id }
func (UserSegment) SystemEntity() bool { return false }

// ############# Entity interface #############

func (UserSegment) PgSqlCreate() {
	db.CreateTable(&UserSegment{})
	db.Model(&UserSegment{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&UserSegment{}).AddForeignKey("email_template_id", "email_templates(id)", "RESTRICT", "CASCADE")
}
func (userSegment *UserSegment) BeforeCreate(scope *gorm.Scope) error {
	userSegment.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&UserSegment{}).Where("account_id = ?",  userSegment.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	userSegment.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (userSegment *UserSegment) AfterFind() (err error) {
	return nil
}

// ######### CRUD Functions ############
func (userSegment UserSegment) create() (Entity, error)  {

	en := userSegment

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

func (UserSegment) get(id uint) (Entity, error) {

	var userSegment UserSegment

	err := userSegment.GetPreloadDb(false,false,true).First(&userSegment, id).Error
	if err != nil {
		return nil, err
	}
	return &userSegment, nil
}
func (userSegment *UserSegment) load() error {

	err := userSegment.GetPreloadDb(false,false,true).First(userSegment, userSegment.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (userSegment *UserSegment) loadByPublicId() error {

	if userSegment.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить UserSegment - не указан  Id"}
	}
	if err := userSegment.GetPreloadDb(false,false, true).First(userSegment, "account_id = ? AND public_id = ?", userSegment.AccountId, userSegment.PublicId).Error; err != nil {
		return err
	}

	return nil
}

func (UserSegment) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return UserSegment{}.getPaginationList(accountId, 0, 100, sortBy, "",nil)
}
func (UserSegment) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{}) ([]Entity, uint, error) {

	userSegments := make([]UserSegment,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&UserSegment{}).GetPreloadDb(true,false,true).Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Preload("EmailTemplate", func(db *gorm.DB) *gorm.DB {
				return db.Select(EmailTemplate{}.SelectArrayWithoutData())
			}).Preload("EmailBox").
			Find(&userSegments, "name ILIKE ? OR subject ILIKE ? OR preview_text ILIKE ?", search,search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&UserSegment{}).
			Where("account_id = ? AND name ILIKE ? OR description ILIKE ? ", accountId, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&UserSegment{}).GetPreloadDb(true,false,true).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Preload("EmailTemplate", func(db *gorm.DB) *gorm.DB {
				return db.Select(EmailTemplate{}.SelectArrayWithoutData())
			}).Preload("EmailBox").
			Find(&userSegments).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&UserSegment{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(userSegments))
	for i := range userSegments {
		entities[i] = &userSegments[i]
	}

	return entities, total, nil
}

func (userSegment *UserSegment) update(input map[string]interface{}) error {

	if err := userSegment.GetPreloadDb(true,false,false).Where(" id = ?", userSegment.Id).
		Omit("id", "account_id","created_at").Updates(input).Error; err != nil {
		return err
	}

	err := userSegment.GetPreloadDb(true,false,false).First(userSegment, userSegment.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (userSegment *UserSegment) delete () error {
	return userSegment.GetPreloadDb(true,true,false).Where("id = ?", userSegment.Id).Delete(userSegment).Error
}
// ######### END CRUD Functions ############

func (userSegment *UserSegment) GetPreloadDb(autoUpdateOff bool, getModel bool, preload bool) *gorm.DB {
	_db := db

	if autoUpdateOff {
		_db = _db.Set("gorm:association_autoupdate", false)
	}
	if getModel {
		_db = _db.Model(userSegment)
	} else {
		_db = _db.Model(&UserSegment{})
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
func (userSegment UserSegment) Execute(data map[string]interface{}) error {
	

	return nil
}
