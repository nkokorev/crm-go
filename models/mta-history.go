package models

import (
	"errors"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"net"
	"strings"
	"time"
)

// История отправок писем в автоматических серия. История хранится какое-то время.
// Храним: факт отправки: чего, кому, когда, число попыток, статус (успех/нет) - для быстрой выборки, открытия / отписки / ip-адрес пользователя (?).
type MTAHistory struct {

	Id     		uint   	`json:"id" gorm:"primaryKey"` // очень большой индекс может быть
	HashId 		string `json:"hash_id" gorm:"type:varchar(12);uniqueIndex;not null;"` // публичный Id для защиты от спама/парсинга
	
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	// ID типа события: какая серия, компания или уведомление
	OwnerId		uint			`json:"owner_id" gorm:"index;type:smallint;not null;"`
	// email_queues, email_campaigns, email_notifications
	OwnerType	EmailSenderType	`json:"owner_type" gorm:"varchar(32);default:'email_queues';not null;"`
	
	// Id пользователя, которому было отправлено письмо серия.
	UserId 	*uint `json:"user_id" gorm:"type:int;"` // при отправке на почту пользователю
	User	User `json:"user"`

	// Кому фактически был отправлен email
	Email 	string `json:"email" gorm:"varchar(255);not null;"`
	
	// Queues: номер шага в очереди, который был совершен. Для статистики серий писем.
	QueueStepId	*uint	`json:"queue_step_id" gorm:"type:smallint;"`

	// Queues: последний ли шаг или промежуточный шаг в цепочке. По нему выборка завершивших серию писем за указанный период времени.
	QueueCompleted 	bool 	`json:"queue_completed" gorm:"type:bool;default:false;"`

	// Какой конкретно шаблон был отправлен. Может пригодиться для истории, кто что кому отправлял.
	EmailTemplateId		*uint	`json:"email_template_id" gorm:"type:int;index;"` // << index для выборки по конкретному письму
	EmailTemplate		EmailTemplate `json:"email_template"`

	// ####### Статистика #######

	// Статистика открытий
	Opens 		uint 		`json:"opens" gorm:"type:smallint;"`
	OpenedAt 	*time.Time  	`json:"opened_at"` // << время 1-го открытия

	// Отписался ли человек. По этому полю будет выборка (для сбора статистики)
	Unsubscribed 	bool 	`json:"unsubscribed" gorm:"type:bool;default:false;"`
	UnsubscribedAt 	*time.Time  `json:"unsubscribed_at"` // << время отписки

	Abuse		bool 		`json:"abuse" gorm:"type:bool;default:false;"`
	AbuseAt 	*time.Time `json:"abuse_at"` // << время 1-й жалобы

	// Ip адрес с которого человек открыл письмо. Может быть полезно для определения GeoLocation.
	NetIp	*net.IP `json:"ipAddr" gorm:"type:cidr;"`

	// Дата+время создания
	CreatedAt time.Time  `json:"created_at"`
}

func (MTAHistory) PgSqlCreate() {
	db.Migrator().CreateTable(&MTAHistory{})
	// db.Model(&MTAHistory{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&MTAHistory{}).AddForeignKey("user_id", "users(id)", "SET NULL", "CASCADE")
	err := db.Exec("ALTER TABLE mta_histories " +
		"ADD CONSTRAINT mta_histories_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT mta_histories_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

	// todo: проработать модель удаления шаблона или связи
	// db.Model(&MTAHistory{}).AddForeignKey("email_template_id", "email_templates(id)", "SET NULL", "CASCADE")

}
func (mtaHistory *MTAHistory) BeforeCreate(tx *gorm.DB) error {
	mtaHistory.Id = 0
	if len(mtaHistory.HashId) < 1 {
		mtaHistory.HashId = strings.ToLower(utils.RandStringBytesMaskImprSrcUnsafe(12, true))
		// todo: тут должен быть какой-то тест на существование

	}
	return nil
}

func (mtaHistory *MTAHistory) AfterCreate(tx *gorm.DB) error {
	// AsyncFire(*Event{}.PaymentCreated(mtaHistory.AccountId, mtaHistory.Id))
	return nil
}
func (mtaHistory *MTAHistory) AfterUpdate(tx *gorm.DB) (err error) {

	// AsyncFire(*Event{}.PaymentUpdated(mtaHistory.AccountId, mtaHistory.Id))

	return nil
}
func (mtaHistory *MTAHistory) AfterDelete(tx *gorm.DB) (err error) {
	// AsyncFire(*Event{}.PaymentDeleted(mtaHistory.AccountId, mtaHistory.Id))
	return nil
}
func (mtaHistory *MTAHistory) AfterFind(tx *gorm.DB) (err error) {
	return nil
}

func (mtaHistory *MTAHistory) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(mtaHistory)
	} else {
		_db = _db.Model(&MTAHistory{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"User","EmailTemplate"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}

// ############# Entity interface #############
func (mtaHistory MTAHistory) GetId() uint { return mtaHistory.Id }
func (mtaHistory *MTAHistory) setId(id uint) { mtaHistory.Id = id }
func (mtaHistory *MTAHistory) setPublicId(publicId uint) { }
func (mtaHistory MTAHistory) GetAccountId() uint { return mtaHistory.AccountId }
func (mtaHistory *MTAHistory) setAccountId(id uint) { mtaHistory.AccountId = id }
func (MTAHistory) SystemEntity() bool { return false }
// ############# Entity interface #############

// ######### CRUD Functions ############
func (mtaHistory MTAHistory) create() (Entity, error)  {
	
	_item := mtaHistory
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}
func (MTAHistory) get(id uint, preloads []string) (Entity, error) {
	var item MTAHistory

	err := item.GetPreloadDb(false, false, preloads).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
func (mtaHistory *MTAHistory) load(preloads []string) error {
	if mtaHistory.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить CartItem - не указан  Id"}
	}

	err := mtaHistory.GetPreloadDb(false, false, preloads).First(mtaHistory, mtaHistory.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (mtaHistory *MTAHistory) loadByPublicId(preloads []string) error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}
func (MTAHistory) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return MTAHistory{}.getPaginationList(accountId, 0, 25, sortBy, "",nil,preload)
}
func (MTAHistory) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	emailQueueHistories := make([]MTAHistory,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&MTAHistory{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).
			Find(&emailQueueHistories, "email ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&MTAHistory{}).GetPreloadDb(false,false,nil).Where(filter).
			Where("account_id = ? AND email ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		
		err := (&MTAHistory{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&emailQueueHistories).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&MTAHistory{}).GetPreloadDb(false,false,nil).Where("account_id = ?", accountId).Where(filter).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(emailQueueHistories))
	for i := range emailQueueHistories {
		entities[i] = &emailQueueHistories[i]
	}

	return entities, total, nil
}
func (MTAHistory) getPaginationListByOwner(owner EmailSender, offset, limit int, sortBy string) ([]MTAHistory, int64, error) {

	emailQueueHistories := make([]MTAHistory,0)
	var total int64

	err := db.Model(&MTAHistory{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ? AND owner_id = ? AND owner_type = ?", owner.GetAccountId(), owner.GetId(), owner.GetType()).
		Find(&emailQueueHistories).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, 0, err
	}

	// Определяем total
	err = db.Model(&MTAHistory{}).Where("account_id = ? AND owner_id = ? AND owner_type = ?", owner.GetAccountId(), owner.GetId(), owner.GetType()).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	return emailQueueHistories, total, nil
}
func (mtaHistory *MTAHistory) update(input map[string]interface{}, preloads []string) error {

	delete(input,"user")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"owner_id","user_id","queue_step_id","email_template_id","opens"}); err != nil {
		return err
	}

	if err := mtaHistory.GetPreloadDb(false, false, nil).Where("id = ?", mtaHistory.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	err := mtaHistory.GetPreloadDb(false,false, preloads).First(mtaHistory, mtaHistory.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (mtaHistory *MTAHistory) delete () error {

	return mtaHistory.GetPreloadDb(true,false,nil).Where("id = ?", mtaHistory.Id).Delete(mtaHistory).Error
}
// ######### END CRUD Functions ############


func (account Account) GetMTAHistoryByHashId(hashId string) (*MTAHistory, error) {
	et := MTAHistory{}

	err := db.First(&et, "account_id = ? AND hash_id = ?", account.Id, hashId).Error
	if err != nil {
		return nil, err
	}
	return &et, nil
}
func (mtaHistory *MTAHistory) UpdateSetUnsubscribeUser(ipV4 string) error {
	return mtaHistory.update(map[string]interface{}{
		"unsubscribed"		: 	true,
		"unsubscribed_at"	: 	time.Now().UTC(),
		// "net_ip"	:	ipV4,
	},nil)
}
func (mtaHistory *MTAHistory) UpdateOpenUser(ipV4 string) error {
	
	input := map[string]interface{} {
		"opens": mtaHistory.Opens + 1,
		// "net_ip"	:	ipV4,
	}
	if mtaHistory.Opens <= 0 {
		input["opened_at"] = time.Now().UTC()
	}

	return mtaHistory.update(input,nil)
}

func (MTAHistory) ExistUserById(userId uint, owner EmailSender) bool {

	// Проверяет историю
	if err := db.Model(&MTAHistory{}).Where("owner_id = ? AND owner_type = ? AND user_id = ?",
		owner.GetId(), owner.GetType(), userId).First(&MTAHistory{}).Error; err != nil {
		return false
	} else {
		return true
	}
}