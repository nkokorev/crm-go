package models

import (
	"errors"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"log"
	"time"
)

// M <> M : email queue  <> email templates
type EmailQueueEmailTemplate struct {

	Id     		uint   	`json:"id" gorm:"primaryKey"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	EmailQueueId	uint	`json:"email_queue_id" gorm:"type:int;"`
	EmailQueue		EmailQueue `json:"-"`
	
	// В работе данное письмо в указанной серии
	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:false;"`
	Step 		uint 	`json:"step" gorm:"type:int;not null;"` // порядок (старый order)

	EmailTemplateId	*uint	`json:"email_template_id" gorm:"type:int;"`
	EmailTemplate	EmailTemplate `json:"email_template"`

	// С какого почтового ящика отправляем
	EmailBoxId		*uint 	`json:"email_box_id" gorm:"type:int;"` // С какого ящика идет отправка
	EmailBox		EmailBox `json:"email_box"`

	// Через сколько запускать письмо в серии. hours / days / week
	DelayTime		time.Duration `json:"delay_time" gorm:"default:0"`// << учитывается только время [0-24]

	// С каким текстом отправляется это сообщение.
	Subject			*string 	`json:"subject" gorm:"type:varchar(255);"` // Тема сообщения, компилируются
	PreviewText		*string 	`json:"preview_text" gorm:"type:varchar(255);"` // Тема сообщения, компилируются


	// График: Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday | weekends, workday
	// Schedule	string `json:"email_template"`     `json:"switch_products"`
	// 1- mondey, workday = 8, weekend = 9, everyday = 10
	// Schedule	postgres.Jsonb `json:"schedule" gorm:"type:JSONB;DEFAULT '{}'::JSONB"`
	Schedule	datatypes.JSON `json:"schedule"`

	// В какое время следует отправлять электронные письма:
	// инста отправка GateStart = GateEnd = null
	// В указанное время: GateStart = <...>, GateEnd = null
	// В указанный промежуток: между GateStart <> GateEnd
	GateStart 	*time.Time `json:"gate_start"`// << учитывается только время [0-24]
	GateEnd		*time.Time `json:"gate_end"`	// << учитывается только время [0-24]

	// Что делать, если Gate не подходит? перенести на 1-24 часа / пропустить письмо и перейти к следующему

	// Что делать, если письмо на паузе - пропускать?
	// SkipIfDisabled bool `json:"skipIfDisabled" gorm:"type:bool;default:true;"`

	// Внутреннее время
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (EmailQueueEmailTemplate) PgSqlCreate() {
	db.Migrator().CreateTable(&EmailQueueEmailTemplate{})
	// db.Model(&EmailQueueEmailTemplate{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&EmailQueueEmailTemplate{}).AddForeignKey("email_queue_id", "email_queues(id)", "CASCADE", "CASCADE")
	// db.Model(&EmailQueueEmailTemplate{}).AddForeignKey("email_template_id", "email_templates(id)", "RESTRICT", "CASCADE")
	err := db.Exec("ALTER TABLE email_queue_email_templates " +
		"ADD CONSTRAINT email_queue_email_templates_notifications_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT email_queue_email_templates_email_queue_id_fkey FOREIGN KEY (email_queue_id) REFERENCES email_queues(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT email_queue_email_templates_email_template_id_fkey FOREIGN KEY (email_template_id) REFERENCES email_templates(id) ON DELETE SET NULL ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (emailQueueEmailTemplate *EmailQueueEmailTemplate) BeforeCreate(tx *gorm.DB) error {
	emailQueueEmailTemplate.Id = 0
	return nil
}

func (emailQueueEmailTemplate *EmailQueueEmailTemplate) AfterCreate(tx *gorm.DB) error {
	// event.AsyncFire(Event{}.PaymentCreated(emailQueueEmailTemplate.AccountId, emailQueueEmailTemplate.Id))
	return nil
}
func (emailQueueEmailTemplate *EmailQueueEmailTemplate) AfterUpdate(tx *gorm.DB) (err error) {

	// event.AsyncFire(Event{}.PaymentUpdated(emailQueueEmailTemplate.AccountId, emailQueueEmailTemplate.Id))

	return nil
}
func (emailQueueEmailTemplate *EmailQueueEmailTemplate) AfterDelete(tx *gorm.DB) (err error) {
	// event.AsyncFire(Event{}.PaymentDeleted(emailQueueEmailTemplate.AccountId, emailQueueEmailTemplate.Id))
	return nil
}
func (emailQueueEmailTemplate *EmailQueueEmailTemplate) AfterFind(tx *gorm.DB) (err error) {

	return nil
}

// ############# Entity interface #############
func (emailQueueEmailTemplate EmailQueueEmailTemplate) GetId() uint { return emailQueueEmailTemplate.Id }
func (emailQueueEmailTemplate *EmailQueueEmailTemplate) setId(id uint) { emailQueueEmailTemplate.Id = id }
func (emailQueueEmailTemplate *EmailQueueEmailTemplate) setPublicId(publicId uint) { }
func (emailQueueEmailTemplate EmailQueueEmailTemplate) GetAccountId() uint { return emailQueueEmailTemplate.AccountId }
func (emailQueueEmailTemplate *EmailQueueEmailTemplate) setAccountId(id uint) { emailQueueEmailTemplate.AccountId = id }
func (EmailQueueEmailTemplate) SystemEntity() bool { return false }
// ############# Entity interface #############

// ######### CRUD Functions ############
func (emailQueueEmailTemplate EmailQueueEmailTemplate) create() (Entity, error)  {
	
	wb := emailQueueEmailTemplate
	
	if err := db.Create(&wb).Error; err != nil {
		return nil, err
	}

	err := wb.GetPreloadDb(false,false, nil).First(&wb, wb.Id).Error
	if err != nil {
		return nil, err
	}

	var entity Entity = &wb

	return entity, nil
}

func (EmailQueueEmailTemplate) get(id uint, preloads []string) (Entity, error) {

	var emailQueueEmailTemplate EmailQueueEmailTemplate

	err := emailQueueEmailTemplate.GetPreloadDb(false,false, preloads).First(&emailQueueEmailTemplate, id).Error
	if err != nil {
		return nil, err
	}
	return &emailQueueEmailTemplate, nil
}
func (EmailQueueEmailTemplate) getByExternalId(externalId string) (*EmailQueueEmailTemplate, error) {
	emailQueueEmailTemplate := EmailQueueEmailTemplate{}

	err := emailQueueEmailTemplate.GetPreloadDb(false,true,nil).First(&emailQueueEmailTemplate, "external_id = ?", externalId).Error
	if err != nil {
		return nil, err
	}
	return &emailQueueEmailTemplate, nil
}
func (emailQueueEmailTemplate *EmailQueueEmailTemplate) load(preloads []string) error {
	if emailQueueEmailTemplate.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить EmailQueueEmailTemplate - не указан  Id"}
	}

	err := emailQueueEmailTemplate.GetPreloadDb(false,false, preloads).First(emailQueueEmailTemplate,emailQueueEmailTemplate.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (emailQueueEmailTemplate *EmailQueueEmailTemplate) loadByPublicId(preloads []string) error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}

func (EmailQueueEmailTemplate) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return EmailQueueEmailTemplate{}.getPaginationList(accountId, 0, 50, sortBy, "",nil,preload)
}

func (EmailQueueEmailTemplate) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	emailQueueEmailTemplates := make([]EmailQueueEmailTemplate,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&EmailQueueEmailTemplate{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&emailQueueEmailTemplates, "gate_start ILIKE ? OR gate_end ILIKE ?", search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&EmailQueueEmailTemplate{}).GetPreloadDb(false,false,nil).
			Where("account_id = ? AND gate_start ILIKE ? OR gate_end ILIKE ?", accountId, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&EmailQueueEmailTemplate{}).GetPreloadDb(false,false,preloads).
			Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).Where(filter).
			Find(&emailQueueEmailTemplates).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&EmailQueueEmailTemplate{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(emailQueueEmailTemplates))
	for i := range emailQueueEmailTemplates {
		entities[i] = &emailQueueEmailTemplates[i]
	}

	return entities, total, nil
}

func (emailQueueEmailTemplate *EmailQueueEmailTemplate) update(input map[string]interface{}, preloads []string) error {

	delete(input,"email_template")
	delete(input,"email_box")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"email_box_id","email_queue_id","email_template_id"}); err != nil {
		return err
	}

	input = utils.FixInputDataTimeVars(input,[]string{"delay_time"})

	if err := emailQueueEmailTemplate.GetPreloadDb(false, false, nil).Where("id = ?", emailQueueEmailTemplate.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	err := emailQueueEmailTemplate.GetPreloadDb(false,false, preloads).First(emailQueueEmailTemplate, emailQueueEmailTemplate.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (emailQueueEmailTemplate *EmailQueueEmailTemplate) delete () error {

	return emailQueueEmailTemplate.GetPreloadDb(true,false,nil).Where("id = ?", emailQueueEmailTemplate.Id).Delete(emailQueueEmailTemplate).Error
}
// ######### END CRUD Functions ############

func (emailQueueEmailTemplate *EmailQueueEmailTemplate) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(emailQueueEmailTemplate)
	} else {
		_db = _db.Model(&EmailQueueEmailTemplate{})
	}

	if autoPreload {
		return _db.Preload("EmailTemplate", func(db *gorm.DB) *gorm.DB {
			return db.Select(EmailTemplate{}.SelectArrayWithoutData())
		}).Preload("EmailQueue").Preload("EmailBox")
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"EmailTemplate","EmailQueue","EmailBox"})

		for _,v := range allowed {
			if v == "EmailTemplate" {
				_db.Preload("EmailTemplate", func(db *gorm.DB) *gorm.DB {
					return db.Select(EmailTemplate{}.SelectArrayWithoutData())
				})
			} else {
				_db.Preload(v)
			}

		}
		return _db
	}

}

func (emailQueueEmailTemplate *EmailQueueEmailTemplate) IsActive() bool {
	return emailQueueEmailTemplate.Enabled
}

func (emailQueueEmailTemplate *EmailQueueEmailTemplate) Validate() error {

	account, err := GetAccount(emailQueueEmailTemplate.AccountId)
	if err != nil {
		return utils.Error{Message: "Ошибка отправления Уведомления - не удается найти аккаунт"}
	}


	// Проверяем тело сообщения (не должно быть пустое)
	if emailQueueEmailTemplate.Subject == nil {
		return utils.Error{Message: "Кампания не может быть запущена, т.к. нет темы сообщения"}
	}
	if emailQueueEmailTemplate.PreviewText == nil {
		emailQueueEmailTemplate.PreviewText = utils.STRp("")
	}

	// Проверяем ключи и загружаем еще раз все данные для отправки сообщения
	if emailQueueEmailTemplate.EmailTemplateId == nil {
		return utils.Error{Message: "Кампания не может быть запущена, т.к. нет установленного шаблона email-сообщения"}
	}
	err = account.LoadEntity(&emailQueueEmailTemplate.EmailTemplate, *emailQueueEmailTemplate.EmailTemplateId,nil)
	if err != nil {
		log.Printf("Ошибка загрузки шаблона email-сообщения для кампании [%v]: %v\n", emailQueueEmailTemplate.Id, err)
		return utils.Error{Message: "Кампания не может быть запущена - ошибка загрузки шаблона email-сообщения"}
	}

	if emailQueueEmailTemplate.EmailBoxId == nil  {
		return utils.Error{Message: "Кампания не может быть запущена, т.к. нет установленного адреса отправителя"}
	}
	err = account.LoadEntity(&emailQueueEmailTemplate.EmailBox, *emailQueueEmailTemplate.EmailBoxId,nil)
	if err != nil {
		log.Printf("Ошибка загрузки адреса отправителя для кампании [%v]: %v\n", emailQueueEmailTemplate.Id, err)
		return utils.Error{Message: "Кампания не может быть запущена - ошибка загрузки адреса отправителя"}
	}

	// Проверяем DKIM подпись
	var webSite WebSite
	err = account.LoadEntity(&webSite, emailQueueEmailTemplate.EmailBox.WebSiteId,nil)
	if err != nil {
		log.Printf("Ошибка загрузки web site отправителя для кампании [%v]: %v\n", emailQueueEmailTemplate.Id, err)
		return utils.Error{Message: "Кампания не может быть запущена - ошибка загрузки адреса отправителя"}
	}

	if err := webSite.ValidateDKIM(); err != nil { return err }

	// Тестовые данные
	// Локальные данные аккаунта, пользователя
	data := make(map[string]interface{})

	data["accountId"] = account.Id
	data["Account"] = account.GetDepersonalizedData() // << хз
	data["userId"] = 0
	data["User"] = User{Id: 1,Name: utils.STRp("TIvan"), Username: utils.STRp("TUsername"), Surname: utils.STRp("TSurname"), Patronymic: utils.STRp("TPatronymic"),
		PhoneRegion: utils.STRp("RU"), Phone: utils.STRp("+79251000000")} // << хз
	data["unsubscribeUrl"] = "/unsubscribe_url"



	viewData := ViewData {
		Subject: *emailQueueEmailTemplate.Subject,
		PreviewText: *emailQueueEmailTemplate.PreviewText,
		Data: data,
		Json: data,
		UnsubscribeURL: "",
		PixelURL: "",
		PixelHTML: "<div></div>",
	}

	return emailQueueEmailTemplate.EmailTemplate.Validate(&viewData)
}
