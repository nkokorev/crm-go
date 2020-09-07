package models

import (
	"database/sql"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

type TaskScheduler struct {
	Id     			uint   	`json:"id" gorm:"primaryKey"`
	PublicId		uint   	`json:"public_id" gorm:"type:int;index;not null;default:1"`
	AccountId 		uint 	`json:"-" gorm:"type:int;index;not null;"`

	// email_campaign_run, email_queues_run, email_notifications_run
	OwnerType	Task	`json:"owner_type" gorm:"varchar(32);not null;"` // << тип события:
	OwnerId		uint	`json:"owner_id" gorm:"type:smallint;not null;"` // ID типа события: id типа

	// Запланированное время выполнения
	ExpectedTimeToStart 	time.Time `json:"expected_time_to_start"`

	// системная ли задача
	IsSystem	bool	`json:"is_system" gorm:"type:bool;default:true"`

	// Результат выполнения: planned / pending / completed / failed / cancelled => планируется / выполняется / выполнена / провалена / отмена
	Status	WorkStatus `json:"status" gorm:"type:varchar(18);default:'planned'"`

	// Причина провала / отмены (необязательный параметр)
	Reason 	*string `json:"reason" gorm:"type:varchar(128);"`

	CreatedAt 		time.Time `json:"created_at"`
	UpdatedAt 		time.Time `json:"updated_at"`
}

type Task = string
const (
	TaskEmailCampaignRun		Task = "email_campaign_run"
	TaskEmailQueueRun			Task = "email_queue_run"
	TaskEmailNotificationRun	Task = "email_notification_run"
)

type WorkStatus = string
const (
	WorkStatusPending		WorkStatus = "pending" // ожидается, рассматривается еще, готовится
	WorkStatusPlanned		WorkStatus = "planned" // запланирована
	WorkStatusActive		WorkStatus = "active"  // выполняется
	WorkStatusPaused		WorkStatus = "paused"
	WorkStatusCompleted		WorkStatus = "completed"
	WorkStatusFailed		WorkStatus = "failed"
	WorkStatusCancelled		WorkStatus = "cancelled"
)

// ############# Entity interface #############
func (taskScheduler TaskScheduler) GetId() uint { return taskScheduler.Id }
func (taskScheduler *TaskScheduler) setId(id uint) { taskScheduler.Id = id }
func (taskScheduler *TaskScheduler) setPublicId(publicId uint) { taskScheduler.PublicId = publicId }
func (taskScheduler TaskScheduler) GetAccountId() uint { return taskScheduler.AccountId }
func (taskScheduler *TaskScheduler) setAccountId(id uint) { taskScheduler.AccountId = id }
func (TaskScheduler) SystemEntity() bool { return false }
// ############# End Entity interface #############

func (TaskScheduler) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&TaskScheduler{}); err != nil {log.Fatal(err)}
	// db.Model(&TaskScheduler{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE task_schedulers ADD CONSTRAINT task_schedulers_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (taskScheduler *TaskScheduler) BeforeCreate(tx *gorm.DB) error {
	taskScheduler.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&TaskScheduler{}).Where("account_id = ?",  taskScheduler.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	taskScheduler.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (taskScheduler *TaskScheduler) AfterFind(tx *gorm.DB) (err error) {
	return nil
}

// ######### CRUD Functions ############
func (taskScheduler TaskScheduler) create() (Entity, error)  {

	en := taskScheduler

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
func (TaskScheduler) get(id uint, preloads []string) (Entity, error) {

	var taskScheduler TaskScheduler

	err := taskScheduler.GetPreloadDb(false,false,preloads).First(&taskScheduler, id).Error
	if err != nil {
		return nil, err
	}
	return &taskScheduler, nil
}
func (taskScheduler *TaskScheduler) load(preloads []string) error {

	err := taskScheduler.GetPreloadDb(false,false,preloads).First(taskScheduler, taskScheduler.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (taskScheduler *TaskScheduler) loadByPublicId(preloads []string) error {
	
	if taskScheduler.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить TaskScheduler - не указан  Id"}
	}
	if err := taskScheduler.GetPreloadDb(false,false, preloads).First(taskScheduler, "account_id = ? AND public_id = ?", taskScheduler.AccountId, taskScheduler.PublicId).Error; err != nil {
		return err
	}

	return nil
}

func (TaskScheduler) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return TaskScheduler{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (TaskScheduler) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	emailCampaigns := make([]TaskScheduler,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&TaskScheduler{}).GetPreloadDb(false,false,preloads).Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&emailCampaigns, "name ILIKE ? OR subject ILIKE ? OR preview_text ILIKE ?", search,search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&TaskScheduler{}).
			Where("account_id = ? AND name ILIKE ? OR subject ILIKE ? OR preview_text ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&TaskScheduler{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&emailCampaigns).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&TaskScheduler{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(emailCampaigns))
	for i := range emailCampaigns {
		entities[i] = &emailCampaigns[i]
	}

	return entities, total, nil
}

func (taskScheduler *TaskScheduler) update(input map[string]interface{}, preloads []string) error {

	utils.FixInputHiddenVars(&input)
	input = utils.FixInputDataTimeVars(input,[]string{"scheduleRun"})
	
	if err := taskScheduler.GetPreloadDb(false,false,nil).Where(" id = ?", taskScheduler.Id).
		Omit("id", "account_id","created_at").Updates(input).Error; err != nil {
		return err
	}

	err := taskScheduler.GetPreloadDb(false,false,preloads).First(taskScheduler, taskScheduler.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (taskScheduler *TaskScheduler) delete () error {
	return taskScheduler.GetPreloadDb(true,false,nil).Where("id = ?", taskScheduler.Id).Delete(taskScheduler).Error
}
// ######### END CRUD Functions ############

func (taskScheduler *TaskScheduler) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(&taskScheduler)
	} else {
		_db = _db.Model(&TaskScheduler{})
	}

	if autoPreload {
		return db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{""})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}

// Выполнения задачи согласно ее контенту
func (taskScheduler TaskScheduler) Execute() error {
	// return errors.New("Тестовая ошибка")

	account, err := GetAccount(taskScheduler.AccountId)
	if err != nil {
		return err
	}

	switch taskScheduler.OwnerType {
	case TaskEmailCampaignRun:

		// 1. Получаем объект кампании
		var emailCampaign EmailCampaign
		err := account.LoadEntity(&emailCampaign, taskScheduler.OwnerId,nil)
		if err != nil {
			log.Printf("Ошибка получения EmailCampaign taskScheduler: %v\n", err)
			return err
		}

		// 2. Запускаем кампанию и ожидаем ответ (результат запуска)
		return emailCampaign.Execute()
		
	default:
		log.Printf("Тип задачи не установлен.. Id: %v Type: %v \n", taskScheduler.Id, taskScheduler.OwnerType)
	}
	
	return nil
}

func (taskScheduler *TaskScheduler) SetStatus(status WorkStatus, ReasonVar... string) error {
	reason := ""
	if len(ReasonVar) > 0 {
		reason = ReasonVar[0]
	}
	return taskScheduler.update(map[string]interface{}{"status":status, "reason":reason},nil)
}
