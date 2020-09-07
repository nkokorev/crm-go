package models

import (
	"errors"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

// Внутренние комментарии менеджеров / ответственных
type Comment struct {
	
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	AccountId 	uint	`json:"account_id" gorm:"index;not null"` // аккаунт-владелец ключа
	UserId 		*uint	`json:"user_id" gorm:"type:int;index;"`
	User		User	`json:"user"`

	Description 		string 	`json:"description" gorm:"type:varchar(255);"` // Описание назначения канала
	ManagersComments 	string `json:"managers_comments" gorm:"type:varchar(255);"`
	
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ############# Entity interface #############
func (comment Comment) GetId() uint { return comment.Id }
func (comment *Comment) setId(id uint) { comment.Id = id }
func (comment *Comment) setPublicId(id uint) { }
func (comment Comment) GetAccountId() uint { return comment.AccountId }
func (comment *Comment) setAccountId(id uint) { comment.AccountId = id }
func (comment Comment) SystemEntity() bool { return false }

// ############# Entity interface #############

func (Comment) PgSqlCreate() {
	db.Migrator().CreateTable(&Comment{})
	// db.Model(&Comment{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&Comment{}).AddForeignKey("user_id", "users(id)", "SET NULL", "CASCADE")
	err := db.Exec("ALTER TABLE comments " +
		"ADD CONSTRAINT order_comments_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT order_comments_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (comment *Comment) BeforeCreate(tx *gorm.DB) error {
	comment.Id = 0
	return nil
}
func (comment *Comment) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(&comment)
	} else {
		_db = _db.Model(&Comment{})
	}

	if autoPreload {
		return db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"User"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}

// ######### CRUD Functions ############
func (comment Comment) create() (Entity, error)  {
	_comment := comment
	if err := db.Create(&_comment).Error; err != nil {
		return nil, err
	}

	if err := _comment.GetPreloadDb(false,false, nil).First(&_comment,_comment.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_comment

	return entity, nil
}

func (Comment) get(id uint, preloads []string) (Entity, error) {

	var comment Comment

	err := (&Comment{}).GetPreloadDb(false,false,preloads).First(&comment, id).Error
	if err != nil {
		return nil, err
	}
	
	return &comment, nil
}
func (comment *Comment) load(preloads []string) error {
	if comment.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить Comment - не указан  Id"}
	}

	err := (&Comment{}).GetPreloadDb(false,false,preloads).First(comment, comment.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (*Comment) loadByPublicId(preloads []string) error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}

func (Comment) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {

	return Comment{}.getPaginationList(accountId, 0,100,sortBy,"",nil,preload)
}

func (Comment) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	comments := make([]Comment,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&Comment{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&comments, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err =(&Comment{}).GetPreloadDb(false,false,nil).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&Comment{}).GetPreloadDb(false,false,preloads).
			Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&comments).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&Comment{}).GetPreloadDb(false,false,nil).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(comments))
	for i := range comments {
		entities[i] = &comments[i]
	}

	return entities, total, nil
}

func (comment *Comment) update(input map[string]interface{}, preloads []string) error {

	delete(input,"user")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"user_id"}); err != nil {
		return err
	}

	if err := comment.GetPreloadDb(false, false, nil).Where("id = ?", comment.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	err := comment.GetPreloadDb(false,false, preloads).First(comment, comment.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (comment *Comment) delete () error {
	return comment.GetPreloadDb(true,false,nil).Where("id = ?", comment.Id).Delete(comment).Error
}
// ######### END CRUD Functions ############


// ########## Work function ############

