package models

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"log"
	"net/mail"
	"time"
)



// Пакет необходимых данных для отправки письма
type EmailPkg struct {
	// From 		mail.Address
	To 			mail.Address
	// Если что-то в отправке пойдет не так - можно будет найти пользователя
	accountId uint
	userId uint

	// тех.данные для отправки
	webSite 	*WebSite
	emailBox 	*EmailBox

	emailSender EmailSender // interface for email-notification, campaign, queue
	// user User

	subject     string // Тема сообщения

	// Шаблон письма
	emailTemplate 	*EmailTemplate

	// Переменные письма письма для компиляции
	viewData		*ViewData

	// Ссылка для отписки (нужна ли, т.к. можно брать из viewData)
	// UnsubscribeUrl string

	// WebSite		WebSite // for DKIM
}

// Types of bounces
type BounceType = string

const(
	// жесткий отскок, когда все совсем плохо
	hardBounced 	BounceType 	= 	"hard"
	
	// мягкий отскок, какая-то временная (возможно) ошибка
	softBounced 	BounceType 	= 	"soft"
)

type MTABounced struct {
	Id     		uint   	`json:"id" gorm:"primary_key"`
	PublicId	uint   	`json:"publicId" gorm:"type:int;index;not null;"`
	AccountId 	uint	`json:"-" gorm:"type:int;index;not null;"`

	// ID типа события: какая серия, компания или уведомление
	OwnerId		uint	`json:"ownerId" gorm:"index;type:smallint;not null;"`
	// email_queues, email_campaigns, email_notifications
	OwnerType	EmailSenderType	`json:"ownerType" gorm:"varchar(32);default:'email_queues';not null;"`

	// Получатель
	UserId		*uint 	`json:"userId" gorm:"type:int;default:null;"`
	User		User	`json:"user"`

	// Почтовый ящик, с которым произошли проблемы
	EmailBoxId	*uint 		`json:"emailBoxId" gorm:"type:int;default:null;"`
	EmailBox 	EmailBox 	`json:"emailBox"`

	// true = soft, false = hard
	SoftBounced bool 	`json:"softBounced" gorm:"type:bool;default:true;"`
	Reason 		string 	`json:"reason" gorm:"type:varchar(255);"`

	CreatedAt 	time.Time 	`json:"createdAt"`
}

// ############# Entity interface #############
func (mtaBounced MTABounced) GetId() uint { return mtaBounced.Id }
func (mtaBounced *MTABounced) setId(id uint) { mtaBounced.Id = id }
func (mtaBounced *MTABounced) setPublicId(publicId uint) { mtaBounced.PublicId = publicId }
func (mtaBounced MTABounced) GetAccountId() uint { return mtaBounced.AccountId }
func (mtaBounced *MTABounced) setAccountId(id uint) { mtaBounced.AccountId = id }
func (MTABounced) SystemEntity() bool { return false }
// ############# End Entity interface #############

func (MTABounced) PgSqlCreate() {
	db.CreateTable(&MTABounced{})
	db.Model(&MTABounced{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&MTABounced{}).AddForeignKey("user_id", "users(id)", "SET NULL", "CASCADE")
	db.Model(&MTABounced{}).AddForeignKey("email_box_id", "email_boxes(id)", "SET NULL", "CASCADE")
}
func (MTABounced) TableName() string {
	return "mta_bounces"
}
func (mtaBounced *MTABounced) BeforeCreate(scope *gorm.Scope) error {
	mtaBounced.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&MTABounced{}).Where("account_id = ?",  mtaBounced.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	mtaBounced.PublicId = 1 + uint(lastIdx.Int64)

	if len(mtaBounced.Reason) > 255 {
		mtaBounced.Reason = mtaBounced.Reason[:254]
	}

	return nil
}
func (mtaBounced *MTABounced) AfterFind() (err error) {
	return nil
}

// ######### CRUD Functions ############
func (mtaBounced MTABounced) create() (Entity, error)  {

	en := mtaBounced

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
func (MTABounced) get(id uint) (Entity, error) {

	var mtaBounced MTABounced

	err := mtaBounced.GetPreloadDb(false,false,true).First(&mtaBounced, id).Error
	if err != nil {
		return nil, err
	}
	return &mtaBounced, nil
}
func (mtaBounced *MTABounced) load() error {

	err := mtaBounced.GetPreloadDb(false,false,true).First(mtaBounced, mtaBounced.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (mtaBounced *MTABounced) loadByPublicId() error {

	if mtaBounced.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить MTABounced - не указан  Id"}
	}
	if err := mtaBounced.GetPreloadDb(false,false, true).
		First(mtaBounced, "account_id = ? AND public_id = ?", mtaBounced.AccountId, mtaBounced.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (MTABounced) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return MTABounced{}.getPaginationList(accountId, 0, 100, sortBy, "",nil)
}
func (MTABounced) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{}) ([]Entity, uint, error) {

	mtaBounces := make([]MTABounced,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&MTABounced{}).GetPreloadDb(true,false,true).Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&mtaBounces, "name ILIKE ? OR subject ILIKE ? OR preview_text ILIKE ?", search,search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&MTABounced{}).
			Where("account_id = ? AND name ILIKE ? OR subject ILIKE ? OR preview_text ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&MTABounced{}).GetPreloadDb(true,false,true).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&mtaBounces).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&MTABounced{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(mtaBounces))
	for i := range mtaBounces {
		entities[i] = &mtaBounces[i]
	}

	return entities, total, nil
}

func (mtaBounced *MTABounced) update(input map[string]interface{}) error {

	input = utils.FixInputHiddenVars(input)
	input = utils.FixInputDataTimeVars(input,[]string{"scheduleRun"})

	if err := mtaBounced.GetPreloadDb(true,false,false).Where(" id = ?", mtaBounced.Id).
		Omit("id", "account_id","created_at").Updates(input).Error; err != nil {
		return err
	}

	err := mtaBounced.GetPreloadDb(true,false,false).First(mtaBounced, mtaBounced.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (mtaBounced *MTABounced) delete () error {
	return mtaBounced.GetPreloadDb(true,true,false).Where("id = ?", mtaBounced.Id).Delete(mtaBounced).Error
}
// ######### END CRUD Functions ############

func (mtaBounced *MTABounced) GetPreloadDb(autoUpdateOff bool, getModel bool, preload bool) *gorm.DB {
	_db := db

	if autoUpdateOff {
		_db = _db.Set("gorm:association_autoupdate", false)
	}
	if getModel {
		_db = _db.Model(mtaBounced)
	} else {
		_db = _db.Model(&MTABounced{})
	}

	if preload {
		return _db.Preload("User").Preload("EmailBox")
		// return _db
	} else {
		return _db
	}
}

// ######### SPECIAL Functions ############




// ######### END OF SPECIAL Functions ############


// обработчик ошибки при отправке письма. Логика по отписке пользователя
func (pkg EmailPkg) bounced(b BounceType, reason string) {

	// 1. Регистрируем отскок в БД
	bounce := MTABounced{
		AccountId: 	pkg.accountId,
		UserId: 	&pkg.userId,
		OwnerId: 	pkg.emailSender.GetId(),
		OwnerType: 	pkg.emailSender.GetType(),
		Reason: 	reason[:255],
		SoftBounced:b == softBounced,
	}
	_,err := bounce.create()
	if err != nil {
		log.Printf("Ошибка создания записи в журнал MTABounced: %v", err)
		return
	}

	user, err := Account{Id: pkg.emailSender.GetAccountId()}.GetUser(pkg.userId)
	if err != nil {
		log.Printf("Ошибка получения user при создании записи в журнал MTABounced: %v", err)
		return
	}

	// 2. Если это мягкий отскок - считаем, сколько их было до этого и принимаем решение - отписывать ли пользователя
	if b == softBounced {
		num, err := bounce.NumSoftByUserId(pkg.userId)
		if err != nil {
			log.Printf("Ошибка подсчета числа soft bounced у пользователя [id=%v]: %v", pkg.userId, err)
			return
		}

		// 5 отскоков, за период в 1 год (по-умолчанию)
		if num >= 5 {
			if err := user.Unsubscribing(); err != nil {
				log.Printf("Ошибка отписки при обновлении user [id=%v] при создании записи в журнал MTABounced: %v", user.Id, err)
				return
			}
		}

	}

	// 3. Если это жесткий отскок - надо отписать пользователя
	if b == hardBounced {

		// отписываем пользователя и указываем время
		if err := user.Unsubscribing(); err != nil {
			log.Printf("Ошибка отписки при обновлении user [id=%v] при создании записи в журнал MTABounced: %v", user.Id, err)
			return
		}
	}

}

// Подсчитывает число мягких отскоков у пользователя в течение 1 года.
func (MTABounced) NumSoftByUserId(userId uint) (uint, error) {

	var total = uint(0)

	err := db.Table("mta_bounces").
		Where("user_id = ? AND soft_bounced = 'true' AND created_at <= ?", userId, time.Now().UTC().AddDate(-1, 0, 0)).Count(&total).Error
	if err != nil {
		return 0, err
	}

	return total, nil
}
