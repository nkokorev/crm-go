package models

import (
	"database/sql"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

type QuestionType = string

const (
	SimpleQuestion 		QuestionType = "simple"
	ProductQuestion 	QuestionType = "product"
	ArticleQuestion 	QuestionType = "article"
	CallBackQuestion 	QuestionType = "callBack"
)

type Question struct {
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	PublicId	uint   	`json:"public_id" gorm:"type:int;index;not null;"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	// => simple / product / article / callBack
	QuestionType	QuestionType `json:"question_type" gorm:"type:varchar(64);default:'simple'"` // тип вопроса
	Status 			WorkStatus 	`json:"status" gorm:"type:varchar(18);default:'pending'"`

	// Id пользователя от которого создан запрос
	UserId 	uint `json:"user_id" gorm:"type:int;not null;"`
	User	User `json:"user"`

	FormName 		*string	`json:"form_name" gorm:"type:varchar(64);"` // Описание что к чему)
	Message 		*string	`json:"message" gorm:"type:varchar(255);"` // Описание что к чему)

	ExpectAnAnswer	bool 	`json:"expect_an_answer" gorm:"type:bool;default:false;"`
	ExpectACallback	bool 	`json:"expect_a_callback" gorm:"type:bool;default:false;"`

	// Откуда пришел IP
	Ipv4 		*string 	`json:"ipv4" gorm:"type:varchar(32);"`

	CreatedAt 	time.Time  `json:"created_at"`
	DeletedAt 	gorm.DeletedAt `json:"-" sql:"index"`
}

// ############# Entity interface #############
func (question Question) GetId() uint { return question.Id }
func (question *Question) setId(id uint) { question.Id = id }
func (question *Question) setPublicId(id uint) {question.PublicId = id}
func (question Question) GetAccountId() uint { return question.AccountId }
func (question *Question) setAccountId(id uint) { question.AccountId = id }
func (Question) SystemEntity() bool { return false }

// ############# Entity interface #############

func (Question) PgSqlCreate() {
	if !db.Migrator().HasTable(&Question{}) {
		if err := db.Migrator().CreateTable(&Question{}); err != nil {log.Fatal(err)}
		err := db.Exec("ALTER TABLE questions ADD CONSTRAINT questions_conditions_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
		if err != nil {
			log.Fatal("Error: ", err)
		}
	}
	/*if db.Migrator().HasTable(&Question{}) {
		err := db.Migrator().DropTable(&Question{})
		if err != nil {
			log.Fatal("Error: ", err)
		}
		// db.Exec("ALTER TABLE questions \n   DROP CONSTRAINT IF EXISTS idx_questions_public_id;").Error
		if err := db.Migrator().CreateTable(&Question{}); err != nil {log.Fatal(err)}
		err = db.Exec("ALTER TABLE questions ADD CONSTRAINT questions_conditions_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
		if err != nil {
			log.Fatal("Error: ", err)
		}
	}*/
	/*if err := db.Migrator().CreateTable(&Question{}); err != nil {log.Fatal(err)}
	err = db.Exec("ALTER TABLE questions ADD CONSTRAINT questions_conditions_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}*/

	/*if err := db.Migrator().AutoMigrate(&Question{}); err != nil {log.Fatal(err)}
	err := db.Exec("ALTER TABLE questions ADD CONSTRAINT questions_conditions_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}*/
}
func (question *Question) BeforeCreate(tx *gorm.DB) error {
	question.Id = 0
	var lastIdx sql.NullInt64

	row := db.Model(&Question{}).Where("account_id = ?",  question.AccountId).
		Select("max(public_id)").Row()
	if row != nil {
		err := row.Scan(&lastIdx)
		if err != nil && err != gorm.ErrRecordNotFound { return err }
	}

	question.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}

// ######### CRUD Functions ############
func (question Question) create() (Entity, error)  {

	_item := question
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}
func (Question) get(id uint, preloads []string) (Entity, error) {

	var item CartItem

	err := item.GetPreloadDb(false, false, preloads).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
func (question *Question) load(preloads []string) error {
	if question.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить Question - не указан  Id"}
	}

	err := question.GetPreloadDb(false, false, preloads).First(question, question.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (question *Question) loadByPublicId(preloads []string) error {

	if question.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить Question - не указан  Id"}
	}

	if err := question.GetPreloadDb(false,false, preloads).
		First(question, "account_id = ? AND public_id = ?", question.AccountId, question.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (Question) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return Question{}.getPaginationList(accountId,0,200,sortBy, "", nil, preload)
}
func (Question) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	webHooks := make([]Question,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&Question{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks, "message ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&Question{}).GetPreloadDb(false, false, nil).
			Where("account_id = ? AND message ILIKE ?", accountId, search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&Question{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Question{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(webHooks))
	for i := range webHooks {
		entities[i] = &webHooks[i]
	}

	return entities, total, nil
}
func (question *Question) update(input map[string]interface{}, preloads []string) error {

	delete(input,"user")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id"}); err != nil {
		return err
	}

	if err := question.GetPreloadDb(false, false, nil).Where("id = ?", question.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	err := question.GetPreloadDb(false,false, preloads).First(question, question.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (question *Question) delete () error {
	return db.Model(Question{}).Where("id = ?", question.Id).Delete(question).Error
}
// ######### END CRUD Functions ############

func (question Question) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(question)
	} else {
		_db = _db.Model(&Question{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"User"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}

// Проверяет на спам
func (question Question) IsSpam(host string) bool {
	var total int64 = 0
	err := db.Model(&Question{}).
		Where("account_id = ? AND created_at >= ? AND ipv4 = ?", question.AccountId, time.Now().Add(-1*time.Hour*24).UTC(), host).
		Count(&total).Error
	if err != nil {return false}
	if total > 3 {
		return true
	}

	total = 0
	err = db.Model(&Question{}).
		Where("account_id = ? AND created_at >= ? AND ipv4 = ?", question.AccountId, time.Now().Add(-1*time.Hour*24*7).UTC(), host).
		Count(&total).Error
	if err != nil {return false}
	if total > 10 {
		return true
	}

	return false
}


func (question *Question) ChangeWorkStatus(status WorkStatus, reason... string) error {

	switch status {
	case WorkStatusPending:
		return question.SetPendingStatus()

	case WorkStatusActive:
		return question.SetActiveStatus()

	case WorkStatusCompleted:
		return question.SetCompletedStatus()

	default:
		return utils.Error{Message: "Статус не опознан"}
	}

}
func (question *Question) updateWorkStatus(status WorkStatus, reason... string) error {

	return question.update(map[string]interface{}{
		"status":	status,
	},nil)
}
func (question *Question) SetPendingStatus() error {

	// Возможен вызов из состояния planned: вернуть на доработку => pending
	if question.Status != WorkStatusPlanned && question.Status != WorkStatusActive {
		reason := "Невозможно установить статус,"
		switch question.Status {
		case WorkStatusPending:
			reason += "т.к. поставка уже в разработке"
		case WorkStatusPaused:
			reason += "т.к. поставка приостановлена"
		case WorkStatusPosting:
			reason += "т.к. поставка в процессе разгрузки"
		case WorkStatusFailed:
			reason += "т.к. поставка завершена с ошибкой"
		case WorkStatusCompleted:
			reason += "т.к. поставка уже завершена"
		case WorkStatusCancelled:
			reason += "т.к. поставка отменена"
		}
		return utils.Error{Message: reason}
	}

	return question.updateWorkStatus(WorkStatusPending)
}

func (question *Question) SetActiveStatus() error {

	// Возможен вызов из состояния planned или paused: запустить поставку => active
	if question.Status != WorkStatusPlanned && question.Status != WorkStatusPaused && question.Status != WorkStatusPending{
		reason := "Невозможно запустить поставку,"
		switch question.Status {
		case WorkStatusActive:
			reason += "т.к. она уже в процессе рассылки"
		case WorkStatusCompleted:
			reason += "т.к. поставка уже завершена"
		case WorkStatusFailed:
			reason += "т.к. поставка завершена с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. поставка отменена"
		}
		return utils.Error{Message: reason}
	}

	// Переводим в состояние "Активна", т.к. все проверки пройдены и можно продолжить ее выполнение
	return question.updateWorkStatus(WorkStatusActive)
}
func (question *Question) SetCompletedStatus() error {

	// Переводим в состояние "Завершена", т.к. все проверки пройдены и можно приостановить поставку
	return question.updateWorkStatus(WorkStatusCompleted)
}
