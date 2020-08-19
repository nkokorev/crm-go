package models

import (
	"database/sql"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"log"
	"time"
)

type TaskScheduler struct {
	Id     			uint   	`json:"id" gorm:"primary_key"`
	PublicId		uint   	`json:"publicId" gorm:"type:int;index;not null;default:1"`
	AccountId 		uint 	`json:"-" gorm:"type:int;index;not null;"`

	// email_campaigns_run, email_queues_run, email_notifications_run
	OwnerType	Task	`json:"ownerType" gorm:"varchar(32);not null;"` // << тип события:
	OwnerId		uint	`json:"ownerId" gorm:"type:smallint;not null;"` // ID типа события: id типа

	// Запланированное время выполнения
	ExpectedTimeToStart 	time.Time `json:"expectedTimeToStart"`

	// системная ли задача
	IsSystem	bool	`json:"isSystem" gorm:"type:bool;default:true"`

	// Результат выполнения: planned / pending / completed / failed / cancelled => планируется / выполняется / выполнена / провалена / отмена
	Status WorkStatus `json:"status" gorm:"type:varchar(18);default:'planned'"`

	// Причина провала / отмены (необязательный параметр)
	Reason string `json:"reason" gorm:"type:varchar(128);default:null"`

	CreatedAt 		time.Time `json:"createdAt"`
	UpdatedAt 		time.Time `json:"updatedAt"`
}

type Task = string
const (
	TaskEmailCampaignRun		Task = "email_campaigns_run"
	TaskEmailQueueRun			Task = "email_queues_run"
	TaskEmailNotificationRun	Task = "email_notifications_run"
)

type WorkStatus = string
const (
	WorkStatusPlanned		WorkStatus = "planned"
	WorkStatusPending		WorkStatus = "pending"
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
	db.CreateTable(&TaskScheduler{})
	db.Model(&TaskScheduler{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}
func (taskScheduler *TaskScheduler) BeforeCreate(scope *gorm.Scope) error {
	taskScheduler.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&TaskScheduler{}).Where("account_id = ?",  taskScheduler.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	taskScheduler.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (taskScheduler *TaskScheduler) AfterFind() (err error) {
	return nil
}

// ######### CRUD Functions ############
func (taskScheduler TaskScheduler) create() (Entity, error)  {

	en := taskScheduler

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

func (TaskScheduler) get(id uint) (Entity, error) {

	var taskScheduler TaskScheduler

	err := taskScheduler.GetPreloadDb(false,false,true).First(&taskScheduler, id).Error
	if err != nil {
		return nil, err
	}
	return &taskScheduler, nil
}
func (taskScheduler *TaskScheduler) load() error {

	err := taskScheduler.GetPreloadDb(false,false,true).First(taskScheduler, taskScheduler.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (taskScheduler *TaskScheduler) loadByPublicId() error {
	
	if taskScheduler.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить TaskScheduler - не указан  Id"}
	}
	if err := taskScheduler.GetPreloadDb(false,false, true).First(taskScheduler, "account_id = ? AND public_id = ?", taskScheduler.AccountId, taskScheduler.PublicId).Error; err != nil {
		return err
	}

	return nil
}

func (TaskScheduler) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return TaskScheduler{}.getPaginationList(accountId, 0, 100, sortBy, "",nil)
}
func (TaskScheduler) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{}) ([]Entity, uint, error) {

	emailCampaigns := make([]TaskScheduler,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&TaskScheduler{}).GetPreloadDb(true,false,true).Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
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

		err := (&TaskScheduler{}).GetPreloadDb(true,false,true).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
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

func (taskScheduler *TaskScheduler) update(input map[string]interface{}) error {

	input = utils.FixInputHiddenVars(input)
	input = utils.FixInputDataTimeVars(input,[]string{"scheduleRun"})
	
	if err := taskScheduler.GetPreloadDb(true,false,false).Where(" id = ?", taskScheduler.Id).
		Omit("id", "account_id","created_at").Updates(input).Error; err != nil {
		return err
	}

	err := taskScheduler.GetPreloadDb(true,false,false).First(taskScheduler, taskScheduler.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (taskScheduler *TaskScheduler) delete () error {
	return taskScheduler.GetPreloadDb(true,true,false).Where("id = ?", taskScheduler.Id).Delete(taskScheduler).Error
}
// ######### END CRUD Functions ############

func (taskScheduler *TaskScheduler) GetPreloadDb(autoUpdateOff bool, getModel bool, preload bool) *gorm.DB {
	_db := db

	if autoUpdateOff {
		_db = _db.Set("gorm:association_autoupdate", false)
	}
	if getModel {
		_db = _db.Model(taskScheduler)
	} else {
		_db = _db.Model(&TaskScheduler{})
	}

	if preload {
		// return _db.Preload("EmailTemplate")
		return _db
	} else {
		return _db
	}
}

// Выполнения задачи 
func (taskScheduler TaskScheduler) Execute() error {
	fmt.Println("Task Execute!")
	// return errors.New("Тестовая ошибка")

	account, err := GetAccount(taskScheduler.AccountId)
	if err != nil {
		return err
	}

	switch taskScheduler.OwnerType {
	case TaskEmailCampaignRun:
		fmt.Println("Запускаем task campaign run!")

		// 1. Получаем кампанию
		var emailCampaign EmailCampaign
		err := account.LoadEntity(&emailCampaign, taskScheduler.OwnerId)
		if err != nil {return err}

		// 2. Запускаем кампанию
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
	return taskScheduler.update(map[string]interface{}{"status":status, "reason":reason})
}
