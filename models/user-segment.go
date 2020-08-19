package models

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type UsersSegment struct {
	Id     			uint   	`json:"id" gorm:"primary_key"`
	PublicId		uint   	`json:"publicId" gorm:"type:int;index;not null;default:1"`
	AccountId 		uint 	`json:"-" gorm:"type:int;index;not null;"`

	// Имя кампании сегмента: 'активные участники', 'только клиенты', 'майские подписчики'
	Name 			string 		`json:"name" gorm:"type:varchar(128);"`

	// true = AND ; false = ANY
	StrictMatching 		bool 	`json:"enabled" gorm:"type:bool;default:false;"`

	// персональные настройки сегмента
	UserSegmentRules []UserSegmentConditions `json:"userSegmentRules" gorm:"many2many:user_segments_user_segment_conditions"`

	// =============   Настройки получателей    ===================
	CreatedAt 		time.Time `json:"createdAt"`
}

// ############# Entity interface #############
func (usersSegment UsersSegment) GetId() uint { return usersSegment.Id }
func (usersSegment *UsersSegment) setId(id uint) { usersSegment.Id = id }
func (usersSegment *UsersSegment) setPublicId(publicId uint) { usersSegment.PublicId = publicId }
func (usersSegment UsersSegment) GetAccountId() uint { return usersSegment.AccountId }
func (usersSegment *UsersSegment) setAccountId(id uint) { usersSegment.AccountId = id }
func (UsersSegment) SystemEntity() bool { return false }

// ############# Entity interface #############

func (UsersSegment) PgSqlCreate() {
	db.CreateTable(&UsersSegment{})
	db.Model(&UsersSegment{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}
func (usersSegment *UsersSegment) BeforeCreate(scope *gorm.Scope) error {
	usersSegment.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&UsersSegment{}).Where("account_id = ?",  usersSegment.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	usersSegment.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (usersSegment *UsersSegment) AfterFind() (err error) {
	return nil
}

// ######### CRUD Functions ############
func (usersSegment UsersSegment) create() (Entity, error)  {

	en := usersSegment

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
func (UsersSegment) get(id uint) (Entity, error) {

	var usersSegment UsersSegment

	err := usersSegment.GetPreloadDb(false,false,true).First(&usersSegment, id).Error
	if err != nil {
		return nil, err
	}
	return &usersSegment, nil
}
func (usersSegment *UsersSegment) load() error {

	err := usersSegment.GetPreloadDb(false,false,true).First(usersSegment, usersSegment.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (usersSegment *UsersSegment) loadByPublicId() error {

	if usersSegment.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить UsersSegment - не указан  Id"}
	}
	if err := usersSegment.GetPreloadDb(false,false, true).First(usersSegment, "account_id = ? AND public_id = ?", usersSegment.AccountId, usersSegment.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (UsersSegment) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return UsersSegment{}.getPaginationList(accountId, 0, 100, sortBy, "",nil)
}
func (UsersSegment) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{}) ([]Entity, uint, error) {

	usersSegments := make([]UsersSegment,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&UsersSegment{}).GetPreloadDb(true,false,true).Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Preload("EmailTemplate", func(db *gorm.DB) *gorm.DB {
				return db.Select(EmailTemplate{}.SelectArrayWithoutData())
			}).Preload("EmailBox").
			Find(&usersSegments, "name ILIKE ? OR subject ILIKE ? OR preview_text ILIKE ?", search,search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&UsersSegment{}).
			Where("account_id = ? AND name ILIKE ? OR description ILIKE ? ", accountId, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&UsersSegment{}).GetPreloadDb(true,false,true).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Preload("EmailTemplate", func(db *gorm.DB) *gorm.DB {
				return db.Select(EmailTemplate{}.SelectArrayWithoutData())
			}).Preload("EmailBox").
			Find(&usersSegments).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&UsersSegment{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(usersSegments))
	for i := range usersSegments {
		entities[i] = &usersSegments[i]
	}

	return entities, total, nil
}
func (usersSegment *UsersSegment) update(input map[string]interface{}) error {

	if err := usersSegment.GetPreloadDb(true,false,false).Where(" id = ?", usersSegment.Id).
		Omit("id", "account_id","created_at").Updates(input).Error; err != nil {
		return err
	}

	err := usersSegment.GetPreloadDb(true,false,false).First(usersSegment, usersSegment.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (usersSegment *UsersSegment) delete () error {
	return usersSegment.GetPreloadDb(true,true,false).Where("id = ?", usersSegment.Id).Delete(usersSegment).Error
}
// ######### END CRUD Functions ############

func (usersSegment *UsersSegment) GetPreloadDb(autoUpdateOff bool, getModel bool, preload bool) *gorm.DB {
	_db := db

	if autoUpdateOff {
		_db = _db.Set("gorm:association_autoupdate", false)
	}
	if getModel {
		_db = _db.Model(usersSegment)
	} else {
		_db = _db.Model(&UsersSegment{})
	}

	if preload {
		// return _db.Preload("")
		return _db
	} else {
		return _db
	}
}

// Получения списка пользователей в сегменте
func (usersSegment UsersSegment) Execute(data map[string]interface{}) error {
	

	return nil
}
