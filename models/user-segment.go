package models

import (
	"database/sql"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)

type UsersSegment struct {
	Id     			uint   	`json:"id" gorm:"primaryKey"`
	PublicId		uint   	`json:"public_id" gorm:"type:int;index;not null;default:1"`
	AccountId 		uint 	`json:"-" gorm:"type:int;index;not null;"`

	// Имя кампании сегмента: 'активные участники', 'только клиенты', 'майские подписчики'
	Name 			string 		`json:"name" gorm:"type:varchar(128);"`

	// true = AND ; false = ANY
	StrictMatching 		bool 	`json:"strict_matching" gorm:"type:bool;default:true;"`

	// персональные настройки сегмента
	UserSegmentRules []UserSegmentCondition `json:"user_segment_rules" gorm:"many2many:user_segments_user_segment_conditions"`
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
	if err := db.Migrator().AutoMigrate(&UsersSegment{}); err != nil {log.Fatal(err)}
	// db.Model(&UsersSegment{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE users_segments ADD CONSTRAINT users_segments_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (usersSegment *UsersSegment) BeforeCreate(tx *gorm.DB) error {
	usersSegment.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&UsersSegment{}).Where("account_id = ?",  usersSegment.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	usersSegment.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (usersSegment *UsersSegment) AfterFind(tx *gorm.DB) (err error) {
	return nil
}

// ######### CRUD Functions ############
func (usersSegment UsersSegment) create() (Entity, error)  {

	_item := usersSegment
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,true, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}
func (UsersSegment) get(id uint, preloads []string) (Entity, error) {

	var usersSegment UsersSegment

	err := usersSegment.GetPreloadDb(false,false,preloads).First(&usersSegment, id).Error
	if err != nil {
		return nil, err
	}
	return &usersSegment, nil
}
func (usersSegment *UsersSegment) load(preloads []string) error {

	err := usersSegment.GetPreloadDb(false,false,preloads).First(usersSegment, usersSegment.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (usersSegment *UsersSegment) loadByPublicId(preloads []string) error {

	if usersSegment.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить UsersSegment - не указан  Id"}
	}
	if err := usersSegment.GetPreloadDb(false,false, preloads).First(usersSegment, "account_id = ? AND public_id = ?", usersSegment.AccountId, usersSegment.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (UsersSegment) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return UsersSegment{}.getPaginationList(accountId, 0, 25, sortBy, "",nil, preload)
}
func (UsersSegment) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	usersSegments := make([]UsersSegment,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&UsersSegment{}).GetPreloadDb(false,false,preloads).
			Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
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
		
		err := (&UsersSegment{}).GetPreloadDb(false,false,preloads).
			Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
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
func (usersSegment *UsersSegment) update(input map[string]interface{}, preloads []string) error {

	delete(input,"user_segment_rules")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id"}); err != nil {
		return err
	}

	if err := usersSegment.GetPreloadDb(true,false,nil).Where(" id = ?", usersSegment.Id).
		Omit("id", "account_id","created_at","public_id").Updates(input).Error; err != nil {
		return err
	}

	err := usersSegment.GetPreloadDb(true,false,preloads).First(usersSegment, usersSegment.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (usersSegment *UsersSegment) delete () error {
	return usersSegment.GetPreloadDb(true,false,nil).Where("id = ?", usersSegment.Id).Delete(usersSegment).Error
}
// ######### END CRUD Functions ############

func (usersSegment *UsersSegment) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(&usersSegment)
	} else {
		_db = _db.Model(&UsersSegment{})
	}

	if autoPreload {
		return db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"Amount","PaymentMethod","Product"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}

// Получения списка пользователей в сегменте
func (usersSegment UsersSegment) Execute(data map[string]interface{}) error {
	

	return nil
}

// Самая важная функция - возвращает пользователей согласно сегменту
func (usersSegment UsersSegment) ChunkUsers(offset int64, limit uint) ([]User, int64, error) {
	// Тут идет сборка всех условий и т.д., но пока парсим просто всех пользователей

	users := make([]User,0)

	// Список пользователей по ролям (выбираем все роли)
	err := db.Table("users").Joins("LEFT JOIN account_users ON account_users.user_id = users.id").
		Select("account_users.account_id, account_users.role_id, users.*").
		Where("account_users.account_id = ? AND users.subscribed = true AND char_length(users.email) > 6", usersSegment.AccountId).
		Offset(int(offset)).Limit(int(limit)).
		Find(&users).Error
	if err != nil {
		log.Printf("ChunkUsers:  %v", err)
		return nil, 0, err
	}

	// Определяем total
	var total = int64(0)
	err = db.Table("users").Joins("LEFT JOIN account_users ON account_users.user_id = users.id").
		Select("account_users.account_id, account_users.role_id, users.*").
		Where("account_users.account_id = ? AND users.subscribed = true AND char_length(users.email) > 6", usersSegment.AccountId).
		Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	return users, total, nil
}