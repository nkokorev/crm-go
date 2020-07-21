package models

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/fatih/structs"
	"github.com/jackc/pgtype"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"github.com/toorop/go-dkim"
	"html/template"
	"log"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

// Template of email body message
type EmailTemplate struct {

	ID     uint   `json:"id" gorm:"primary_key"`
	HashID string `json:"hashId" gorm:"type:varchar(12);unique_index;not null;"` // публичный ID для защиты от спама/парсинга
	AccountID uint `json:"-" gorm:"type:int;index;not null;"`

	Name 		string	`json:"name" gorm:"type:varchar(255);not null"` // inside name of mail
	Description	string 	`json:"description" gorm:"type:varchar(255);default:''"` // краткое назначение письма
	PreviewText string 	`json:"previewText" gorm:"type:varchar(255);default:''"` // превью текст может использоваться, да

	Data string `json:"data, omitempty" gorm:"type:text;"` // сам шаблон письма

	Public bool `json:"public" gorm:"type:bool;"` // показывать ли на домене public

	// User *User `json:"-" sql:"-"` // Пользователь, который получит сообщение
	Json pgtype.JSON `json:"json" gorm:"type:json;default:'{\"Example\":\"Тестовые данные в формате json\"}'"`

	// GORM vars
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	// DeletedAt *time.Time `json:"deletedAt" sql:"index"`
}

func (EmailTemplate) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&EmailTemplate{})
	db.Model(&EmailTemplate{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}

// ############# Entity interface #############
func (emailTemplate EmailTemplate) getId() uint { return emailTemplate.ID }
func (emailTemplate *EmailTemplate) setId(id uint) { emailTemplate.ID = id }
func (emailTemplate EmailTemplate) GetAccountId() uint { return emailTemplate.AccountID }
func (emailTemplate *EmailTemplate) setAccountId(id uint) { emailTemplate.AccountID = id }
func (EmailTemplate) systemEntity() bool { return false }
func (emailTemplate EmailTemplate) GetData() string { return emailTemplate.Data }
// ############# Entity interface #############

func (et *EmailTemplate) BeforeCreate(scope *gorm.Scope) error {
	et.ID = 0
	et.HashID = strings.ToLower(utils.RandStringBytesMaskImprSrcUnsafe(12, true))
	et.CreatedAt = time.Now().UTC()

	return nil
}

// ########### CRUD FUNCTIONAL #########
func (emailTemplate EmailTemplate) create() (Entity, error) {

	et := emailTemplate

	if err := db.Create(&et).Error; err != nil {
		return nil, err
	}
	var entity Entity = &et

	return entity, nil
}

func (EmailTemplate) get(id uint) (Entity, error) {

	var et EmailTemplate

	err := db.First(&et, id).Error
	if err != nil {
		return nil, err
	}
	return &et, nil
}
func (et *EmailTemplate) load() error {

	err := db.First(et).Error
	if err != nil {
		return err
	}
	return nil
}

func (EmailTemplate) getByHashId(hashId string) (*EmailTemplate, error) {
	et := EmailTemplate{}

	err := db.First(&et, "hash_id = ?", hashId).Error
	if err != nil {
		return nil, err
	}
	return &et, nil
}

func (EmailTemplate) getList(accountId uint, sortBy string) ([]Entity, uint, error) {

	emailTemplates := make([]EmailTemplate,0)
	var total uint

	err := db.Model(&EmailTemplate{}).Limit(1000).Order(sortBy).Where( "account_id = ?", accountId).
		Select(EmailTemplate{}.SelectArrayWithoutData()).Find(&emailTemplates).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, 0, err
	}

	// Определяем total
	err = db.Model(&EmailTemplate{}).Where("account_id = ?", accountId).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(emailTemplates))
	for i,_ := range emailTemplates {
		entities[i] = &emailTemplates[i]
	}

	return entities, total, nil
}
func (EmailTemplate) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

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

func (et *EmailTemplate) update(input map[string]interface{}) error {
	// return db.Model(&EmailTemplate{}).Where("id = ?", et.ID).Omit("id", "account_id").Update(input).Error
	return db.Model(et).Where("id = ?", et.ID).Omit("id", "account_id").Update(input).Error
}


func (et EmailTemplate) delete () error {
	return db.Model(EmailTemplate{}).Where("id = ?", et.ID).Delete(et).Error
}
// ########### ACCOUNT FUNCTIONAL ###########

func (account Account) EmailTemplateGetByHashID(hashId string) (*EmailTemplate, error) {
	et, err := (EmailTemplate{}).getByHashId(hashId)
	if err != nil {
		return nil, err
	}

	if et.AccountID != account.ID {
		return nil, errors.New("Шаблон принадлежит другому аккаунту")
	}

	return et, nil
}
func (Account) EmailTemplateGetSharedByHashID(hashId string) (*EmailTemplate, error) {
	et, err := (EmailTemplate{}).getByHashId(hashId)
	if err != nil {
		return nil, err
	}

	if !et.Public {
		return nil, errors.New("Шаблон не расшарен для просмотра через web")
	}

	return et, nil
}
// ########### END OF ACCOUNT FUNCTIONAL ###########

// Подготавливает данные для отправки обезличивая их
func (et EmailTemplate) PrepareViewData(data map[string]interface{}) (*ViewData, error) {

	// 1. Готовим JSON
	jsonMap := make(map[string]interface{})
	err := et.Json.AssignTo(&jsonMap)
	if err != nil {
		return nil, errors.New("Json data not valid")
	}

	return &ViewData{
		TemplateName: et.Name, // ? надо ли?
		PreviewText: et.PreviewText,
		Data: data,
		Json: jsonMap,
	}, nil
}

// Возвращает тело письма в формате string в кодировке HTML, учитывая переменные в T[map]
func (et EmailTemplate) GetHTML(viewData *ViewData) (html string, err error) {

	body := new(bytes.Buffer)

	tmpl, err := template.New(et.Name).Parse(et.Data)
	if err != nil {
		return "", err
	}

	err = tmpl.Execute(body, viewData)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Ошибка email-шаблона: %s\r", err))
	}
	
	return body.String(), nil
}

// user - получатель письма
func (et EmailTemplate) Send(from EmailBox, user User, subject string) error {

	// Принадлежность пользователя к аккаунту не проверяем, т.к. это пофигу
	// user - получатель письма, письмо уйдет на user.Email

	// Формируем данные для сборки шаблона
	// vData := ViewData{et, user, json}
	vData, err := et.PrepareViewData(nil)

	// 1. Получаем html из email'а
	html, err := et.GetHTML(vData)
	if err != nil {
		return err
	}
	
	// 2. Отправляем
	headers := make(map[string]string)

	address := from.GetMailAddress()
	headers["From"] = address.String()
	headers["To"] = user.Email
	headers["Subject"] = subject

	headers["MIME-Version"] = "1.0" // имя SMTP сервера
	headers["Content-Type"] = "text/html; charset=UTF-8"
	headers["Content-Transfer-Encoding"] = "quoted-printable" // имя SMTP сервера
	headers["Feedback-ID"] = "1324078:20488:trust:54854"
	// Идентификатор представляет собой 32-битное число в диапазоне от 1 до 2147483647, либо строку длиной до 40 символов, состоящую из латинских букв, цифр и символов ".-_".
	headers["Message-ID"] = "1001" // номер сообщения (внутренний номер)
	headers["Received"] = "RatusCRM"
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

	_, host := split(user.Email)
	
	privRSAKey := from.WebSite.DKIMPrivateRSAKey

	options := dkim.NewSigOptions()
	options.PrivateKey = []byte(privRSAKey)
	//options.Domain = "rtcrm.ru"
	options.Domain = from.WebSite.Hostname
	options.Selector = "dk1"
	options.SignatureExpireIn = 0
	options.BodyLength = 50
	//options.Headers = []string{"from", "date", "mime-version", "received", "received"}
	options.Headers = GetHeaderKeys(headers)
	options.AddSignatureTimestamp = false
	options.Canonicalization = "relaxed/relaxed"

	//////////////////////

	email := []byte(message)
	if err := dkim.Sign(&email, options); err != nil {
		return errors.New("Cant sign")
	}

	mx, err := net.LookupMX(host)
	if err != nil {
		log.Fatal("Не найдена MX-запись")
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
		log.Fatalf("client.StartTLS fail: %v", err)
	}

	// from
	// err = client.Mail(from.GetMailAddress().Address)
	err = client.Mail("userId.abuse.@ratuscrm.com")
	if err != nil {
		log.Fatal("Почтовый адрес не может принять почту")
	}

	err = client.Rcpt(user.Email)
	if err != nil {
		log.Fatal("Похоже, почтовый адрес не сущесвует")
	}

	wc, err := client.Data()
	if err != nil {
		log.Fatal(err)
	}

	_, err = wc.Write(email)
	if err != nil {
		log.Fatal(err)
	}

	err = wc.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Send the QUIT command and close the connection.
	err = client.Quit()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func (et EmailTemplate) SendMail(from EmailBox, toEmail string, subject string, vData *ViewData) error {

	if from.WebSite.ID <1 {
		log.Println("EmailTemplate: Не удалось определить WebSite")
		return utils.Error{Message: "Не удалось определить WebSite"}
	}

	// Принадлежность пользователя к аккаунту не проверяем, т.к. это пофигу
	// user - получатель письма, письмо уйдет на user.Email

	// Формируем данные для сборки шаблона
	// vData, err := et.PrepareViewData(data)

	// 1. Получаем html из email'а
	html, err := et.GetHTML(vData)
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
	headers["Feedback-ID"] = "1324078:20488:trust:54854"
	// Идентификатор представляет собой 32-битное число в диапазоне от 1 до 2147483647, либо строку длиной до 40 символов, состоящую из латинских букв, цифр и символов ".-_".
	headers["Message-ID"] = "1001" // номер сообщения (внутренний номер)
	headers["Received"] = "RatusCRM"
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
		fmt.Println(err)
		return errors.New("Cant sign async")
	}

	mx, err := net.LookupMX(host)
	if err != nil {
		fmt.Println("Email: ", toEmail)
		fmt.Println(err)
		log.Fatal("Не найдена MX-запись")
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

func (et EmailTemplate) SendChannel(emailBox EmailBox, toEmail string, subject string, inputData map[string]interface{}) error {
	
	account, _ := GetAccount(et.AccountID)
	data, err := et.PrepareViewData(inputData)
	if err != nil || data == nil {
		return errors.New("Ошибка сбора данных для шаблона")
	}

	pkg := EmailPkg {
		From: 	emailBox.GetMailAddress(),
		To: 	mail.Address{Address: toEmail},
		Subject: 	subject,
		WebSite: 	emailBox.WebSite,
		EmailTemplate: et,
		ViewData:  	*data,
		Account:	*account,
	}

	SendEmail(pkg)

	/*for i := 0; i <= 20 ; i++ {
		SendEmail(pkg)
	}

	fmt.Println("Все сообщения в очередь отправлены")*/

	return nil

	// var pkg EmailPkg
	// size := int(unsafe.Sizeof(pkg))
	// fmt.Printf("Size: %d\n", size) // 16 байт   | 112 байт с аккаунтом | 255 с данными

	// return nil

	
	vData, err := et.PrepareViewData(nil)

	// 1. Получаем html из email'а
	html, err := et.GetHTML(vData)
	if err != nil {
		return err
	}

	// 2. Собираем хедеры
	headers := make(map[string]string)

	address := emailBox.GetMailAddress()
	headers["From"] = address.String()
	headers["To"] = toEmail
	headers["Subject"] = subject

	// Статичные хедеры
	headers["MIME-Version"] = "1.0" // имя SMTP сервера
	headers["Content-Type"] = "text/html; charset=UTF-8"
	headers["Content-Transfer-Encoding"] = "quoted-printable" // имя SMTP сервера
	headers["Feedback-ID"] = "1324078:20488:trust:54854"
	// Идентификатор представляет собой 32-битное число в диапазоне от 1 до 2147483647, либо строку длиной до 40 символов, состоящую из латинских букв, цифр и символов ".-_".
	headers["Message-ID"] = "1001" // номер сообщения (внутренний номер)
	headers["Received"] = "RatusCRM"
	// headers["Return-Path"] = "<smtp@rus-marketing.ru>"

	// Create message body with headers
	message := ""
	for k,v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	var buf bytes.Buffer // body of message
	w := quotedprintable.NewWriter(&buf)
	_, err = w.Write([]byte(html))
	if err != nil {
		return nil
	}

	if err = w.Close(); err != nil {
		return nil
	}

	message += "\r\n" + buf.String()

	_, host := split(toEmail) // получаем хост, на который нужно совершить отправку данных

	privRSAKey := emailBox.WebSite.DKIMPrivateRSAKey

	options := dkim.NewSigOptions()
	options.PrivateKey = []byte(privRSAKey)
	options.Domain = emailBox.WebSite.Hostname
	options.Selector = "dk1"
	options.SignatureExpireIn = 0
	options.BodyLength = 50
	options.Headers = GetHeaderKeys(headers)
	options.AddSignatureTimestamp = false
	options.Canonicalization = "relaxed/relaxed"

	//////////////////////

	email := []byte(message)
	if err := dkim.Sign(&email, options); err != nil {
		return errors.New("Cant sign")
	}

	mx, err := net.LookupMX(host)
	if err != nil {
		log.Fatal("Не найдена MX-запись")
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
		log.Fatalf("client.StartTLS fail: %v", err)
	}

	// from
	// err = client.Mail(from.GetMailAddress().Address)
	err = client.Mail("userId.abuse.@ratuscrm.com")
	if err != nil {
		log.Fatal("Почтовый адрес не может принять почту")
	}

	err = client.Rcpt(toEmail)
	if err != nil {
		log.Fatal("Похоже, почтовый адрес не сущесвует")
	}

	wc, err := client.Data()
	if err != nil {
		log.Fatal(err)
	}

	_, err = wc.Write(email)
	if err != nil {
		log.Fatal(err)
	}

	err = wc.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Send the QUIT command and close the connection.
	err = client.Quit()
	if err != nil {
		log.Fatal(err)
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
