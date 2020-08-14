package models

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nkokorev/crm-go/utils"
	"github.com/toorop/go-dkim"
	"html/template"
	"log"
	"mime/quotedprintable"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// Template of email body message
type EmailTemplate struct {

	Id     		uint   	`json:"id" gorm:"primary_key"`
	PublicId	uint   	`json:"publicId" gorm:"type:int;index;not null;default:1"`
	HashId 		string 	`json:"hashId" gorm:"type:varchar(12);unique_index;not null;"` // публичный Id для защиты от спама/парсинга
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	Name 		string	`json:"name" gorm:"type:varchar(255);not null"` // inside name of mail
	Description	string 	`json:"description" gorm:"type:varchar(255);default:''"` // краткое назначение письма
	PreviewText string 	`json:"previewText" gorm:"type:varchar(255);default:''"` // превью текст может использоваться, да

	HTMLData string `json:"htmlData" gorm:"type:text;"` // сам шаблон письма

	Public bool `json:"public" gorm:"type:bool;"` // показывать ли на домене public

	// User *User `json:"-" sql:"-"` // Пользователь, который получит сообщение
	// Json pgtype.JSON `json:"json" gorm:"type:json;default:'{\"Example\":\"Тестовые данные в формате json\"}'"`
	JsonData postgres.Jsonb `json:"jsonData" gorm:"type:JSONB;DEFAULT '{}'::JSONB"`

	// GORM vars
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	// Шаблоны не удаляемы теперь для MTAHistory
	DeletedAt *time.Time `json:"deletedAt" sql:"index"`
}

func (EmailTemplate) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&EmailTemplate{})
	db.Model(&EmailTemplate{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}
func (emailTemplate *EmailTemplate) BeforeCreate(scope *gorm.Scope) error {
	emailTemplate.Id = 0
	emailTemplate.HashId = strings.ToLower(utils.RandStringBytesMaskImprSrcUnsafe(12, true))

	// PublicId
	lastIdx := uint(0)
	var eq EmailTemplate

	err := db.Where("account_id = ?", emailTemplate.AccountId).Select("public_id").Last(&eq).Error
	if err != nil && err != gorm.ErrRecordNotFound { return err}
	if err == gorm.ErrRecordNotFound {
		lastIdx = 0
	} else {
		lastIdx = eq.PublicId
	}
	emailTemplate.PublicId = lastIdx + 1
	
	return nil
}

// ############# Entity interface #############
func (emailTemplate EmailTemplate) GetId() uint { return emailTemplate.Id }
func (emailTemplate *EmailTemplate) setId(id uint) { emailTemplate.Id = id }
func (emailTemplate *EmailTemplate) setPublicId(publicId uint) { emailTemplate.PublicId = publicId}
func (emailTemplate EmailTemplate) GetAccountId() uint { return emailTemplate.AccountId }
func (emailTemplate *EmailTemplate) setAccountId(id uint) { emailTemplate.AccountId = id }
func (EmailTemplate) SystemEntity() bool { return false }
func (emailTemplate EmailTemplate) GetData() string { return emailTemplate.HTMLData }
// ############# Entity interface #############


// ########### CRUD FUNCTIONAL #########
func (emailTemplate EmailTemplate) create() (Entity, error) {

	et := emailTemplate

	if err := db.Create(&et).Error; err != nil {
		return nil, err
	}

	err := et.GetPreloadDb(false,false, true).First(&et, et.Id).Error
	if err != nil {
		return nil, err
	}

	var newItem Entity = &et

	return newItem, nil
}

func (EmailTemplate) get(id uint) (Entity, error) {

	var emailTemplate EmailTemplate

	err := db.First(&emailTemplate, id).Error
	if err != nil {
		return nil, err
	}
	return &emailTemplate, nil
}
func (emailTemplate *EmailTemplate) load() error {

	err := db.First(emailTemplate, emailTemplate.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (emailTemplate *EmailTemplate) loadByPublicId() error {
	if emailTemplate.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить EmailNotification - не указан  Id"}
	}
	if err := emailTemplate.GetPreloadDb(false,false, true).
		First(emailTemplate, "account_id = ? AND public_id = ?", emailTemplate.AccountId, emailTemplate.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (EmailTemplate) getByHashId(hashId string) (*EmailTemplate, error) {
	emailTemplate := EmailTemplate{}

	err := db.First(&emailTemplate, "hash_id = ?", hashId).Error
	if err != nil {
		return nil, err
	}
	return &emailTemplate, nil
}
func (EmailTemplate) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return (EmailTemplate{}).getPaginationList(accountId, 0, 25, sortBy, "", nil)
}
func (EmailTemplate) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{}) ([]Entity, uint, error) {

	emailTemplates := make([]EmailTemplate,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&EmailTemplate{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Select(EmailTemplate{}.SelectArrayWithoutData()).
			Find(&emailTemplates, "hash_id ILIKE ? OR name ILIKE ? OR description ILIKE ? OR preview_text ILIKE ?", search,search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&EmailTemplate{}).
			Where("account_id = ? AND hash_id ILIKE ? OR name ILIKE ? OR description ILIKE ? OR preview_text ILIKE ?", accountId, search,search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&EmailTemplate{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Select(EmailTemplate{}.SelectArrayWithoutData()).
			Find(&emailTemplates).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&EmailTemplate{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(emailTemplates))
	for i,_ := range emailTemplates {
		entities[i] = &emailTemplates[i]
	}

	return entities, total, nil
}

func (emailTemplate *EmailTemplate) update(input map[string]interface{}) error {
	// return db.Model(&EmailTemplate{}).Where("id = ?", emailTemplate.Id).Omit("id", "account_id").Update(input).Error

	input = utils.FixJSONB_String(input, []string{"jsonData"})

	return db.Model(emailTemplate).Omit("id", "account_id").Updates(input).Error
}
func (emailTemplate *EmailTemplate) delete () error {
	return db.Model(EmailTemplate{}).Where("id = ?", emailTemplate.Id).Delete(emailTemplate).Error
}
// ########### ACCOUNT FUNCTIONAL ###########

func (account Account) EmailTemplateGetByHashId(hashId string) (*EmailTemplate, error) {
	emailTemplate, err := (EmailTemplate{}).getByHashId(hashId)
	if err != nil {
		return nil, err
	}

	if emailTemplate.AccountId != account.Id {
		return nil, errors.New("Шаблон принадлежит другому аккаунту")
	}

	return emailTemplate, nil
}
func (Account) EmailTemplateGetSharedByHashId(hashId string) (*EmailTemplate, error) {
	emailTemplate, err := (EmailTemplate{}).getByHashId(hashId)
	if err != nil {
		return nil, err
	}

	if !emailTemplate.Public {
		return nil, errors.New("Шаблон не расшарен для просмотра через web")
	}

	return emailTemplate, nil
}
// ########### END OF ACCOUNT FUNCTIONAL ###########

// Подготавливает данные для отправки обезличивая их
func (emailTemplate EmailTemplate) PrepareViewData(subject, previewText string, data map[string]interface{}, pixelURL string, unsubscribeUrl *string) (*ViewData, error) {

	// 1. Готовим JSON
	// WORK OLD !!!
	/*jsonMap := make(map[string]interface{})
	err := emailTemplate.Json.AssignTo(&jsonMap)
	if err != nil {
		return nil, errors.New("Json data not valid")
	}*/
	unsubUrl := ""
	if unsubscribeUrl != nil {
		unsubUrl = *unsubscribeUrl
	}

	jsonMap := make(map[string]interface{})
	jsonMap = utils.ParseJSONBToMapString(emailTemplate.JsonData)
	
	return &ViewData{
		Subject: subject,
		PreviewText: previewText,
		Data: data,
		Json: jsonMap,
		UnsubscribeURL: unsubUrl,
		PixelURL: pixelURL,
		PixelHTML: emailTemplate.GetPixelHTML(pixelURL),
	}, nil
}

// Возвращает тело письма в формате string в кодировке HTML, учитывая переменные в T[map]
func (emailTemplate EmailTemplate) GetHTML(viewData *ViewData) (html string, err error) {

	body := new(bytes.Buffer)

	tmpl, err := template.New(emailTemplate.Name).Parse(emailTemplate.HTMLData)
	if err != nil {
		return "", err
	}

	err = tmpl.Execute(body, viewData)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Ошибка email-шаблона: %s\r", err))
	}
	
	return body.String(), nil
}

func (emailTemplate EmailTemplate) SendMail(from EmailBox, toEmail string, subject string, vData *ViewData, unsubscribeUrl string) error {
	
	// fmt.Println("unsubscribeUrl: ", unsubscribeUrl)
	// fmt.Println("Типа отослали")
	// return nil
	// return errors.New("sdds")

	if from.WebSite.Id < 1 {
		log.Println("EmailTemplate: Не удалось определить WebSite")
		return utils.Error{Message: "Не удалось определить WebSite"}
	}

	// Принадлежность пользователя к аккаунту не проверяем, т.к. это пофигу
	// user - получатель письма, письмо уйдет на user.Email

	// Формируем данные для сборки шаблона
	// vData, err := emailTemplate.PrepareViewData(data)

	// 1. Получаем html из email'а
	html, err := emailTemplate.GetHTML(vData)
	if err != nil {
		return err
	}

	// 2. Отправляем
	headers := make(map[string]string)

	address := from.GetMailAddress()
	headers["From"] = address.String()
	headers["To"] = toEmail
	headers["Subject"] = subject

	headers["MIME-Version"] = "1.0" // имя SMTP сервера
	headers["Content-Type"] = "text/html; charset=UTF-8"
	headers["Content-Transfer-Encoding"] = "quoted-printable" // имя SMTP сервера
	headers["Feedback-Id"] = "1324078:20488:trust:54854"
	// Идентификатор представляет собой 32-битное число в диапазоне от 1 до 2147483647, либо строку длиной до 40 символов, состоящую из латинских букв, цифр и символов ".-_".
	headers["Message-Id"] = "1001" // номер сообщения (внутренний номер) <<< вот тут не плохо бы сгенерировать какой-нибудь ключ
	headers["Received"] = "RatusCRM"
	headers["List-Unsubscribe"] = unsubscribeUrl
	headers["List-Unsubscribe-Post"] = "List-Unsubscribe=One-Click"
	// headers["Return-Path"] = "<smtp@rus-marketing.ru>"

	// Setup message body
	message := ""
	for k,v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	var buf bytes.Buffer
	w := quotedprintable.NewWriter(&buf)
	_, err = w.Write([]byte(html))
	if err != nil {
		return nil
	}

	if err = w.Close(); err != nil {
		return nil
	}

	message += "\r\n" + buf.String()

	_, host := split(toEmail)

	privRSAKey := from.WebSite.DKIMPrivateRSAKey

	options := dkim.NewSigOptions()
	options.PrivateKey = []byte(privRSAKey)
	//options.Domain = "rtcrm.ru"
	options.Domain = from.WebSite.Hostname
	options.Selector = from.WebSite.DKIMSelector // dk1
	options.SignatureExpireIn = 0
	options.BodyLength = 50
	//options.Headers = []string{"from", "date", "mime-version", "received", "received"}
	options.Headers = GetHeaderKeys(headers)
	options.AddSignatureTimestamp = false
	options.Canonicalization = "relaxed/relaxed"

	//////////////////////

	email := []byte(message)
	if err := dkim.Sign(&email, options); err != nil {
		return errors.New("Cant sign async")
	}

	mx, err := net.LookupMX(host)
	if err != nil {
		// fmt.Println("Email: ", toEmail)
		// fmt.Println(err)
		log.Println("Не найдена MX-запись")
	}

	//addr := fmt.Sprintf("%s:%d", mx[0].Host, 25)
	addr := fmt.Sprintf("%s:%d", mx[0].Host, 25)

	client, err := smtp.Dial(addr)
	if err != nil {
		log.Fatalf("DialTimeout fail: %v", mx[0].Host)
	}

	if err = client.StartTLS(&tls.Config {
		InsecureSkipVerify: true,
		ServerName: host,
	}); err != nil {
		log.Printf("client.StartTLS fail: %v", err)
	}

	// from
	// err = client.Mail(from.GetMailAddress().Address)
	err = client.Mail("user-21.abuse.@ratuscrm.com")
	if err != nil {
		log.Println("Почтовый адрес не может принять почту")
	}

	err = client.Rcpt(toEmail)
	if err != nil {
		log.Println("Похоже, почтовый адрес не существует")
	}

	wc, err := client.Data()
	if err != nil {
		log.Println(err)
	}

	_, err = wc.Write(email)
	if err != nil {
		log.Println(err)
	}

	err = wc.Close()
	if err != nil {
		log.Println(err)
	}

	// Send the QUIT command and close the connection.
	err = client.Quit()
	if err != nil {
		log.Println(err)
	}

	return nil
}

func split(email string) (account, host string) {
	i := strings.LastIndexByte(email, '@')
	account = email[:i]
	host = email[i+1:]
	return
}

func GetHeaderKeys(e map[string]string) (headers []string) {
	for key := range e {
		headers = append(headers, key)
	}
	return headers
}

func (EmailTemplate) SelectArrayWithoutData() []string {
	fields := structs.Names(&EmailTemplate{}) //.(map[string]string)
	fields = utils.RemoveKey(fields, "Data")
	return utils.ToLowerSnakeCaseArr(fields)
}

func (emailTemplate *EmailTemplate) GetPreloadDb(autoUpdateOff bool, getModel bool, preload bool) *gorm.DB {
	_db := db

	if autoUpdateOff {
		_db = _db.Set("gorm:association_autoupdate", false)
	}
	if getModel {
		_db = _db.Model(emailTemplate)
	} else {
		_db = _db.Model(&EmailTemplate{})
	}

	if preload {
		// Preload...
		return _db
	} else {
		return _db
	}
}