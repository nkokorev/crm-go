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
	"net/smtp"
	"strings"
	"time"
)

// Template of email body message
type EmailTemplate struct {

	ID     uint   `json:"-" gorm:"primary_key"`
	HashID string `json:"hashId" gorm:"type:varchar(12);unique_index;not null;"` // публичный ID для защиты от спама/парсинга
	AccountID uint `json:"-" gorm:"type:int;index;not_null;"`

	Public bool `json:"public" gorm:"type:bool;default:true;"` // показывать ли на домене public

	Name string `json:"name" gorm:"type:varchar(255);not_null"` // inside name of mail
	Code string `json:"code, omitempty" gorm:"type:text;"` // сам шаблон письма

	// Data
	// User *User `json:"-" sql:"-"` // Пользователь, который получит сообщение
	Json pgtype.JSON `json:"json" gorm:"type:json;default:'{\"Example\":\"Тестовые данные в формате json\"}'"`

	// GORM vars
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	// DeletedAt *time.Time `json:"deletedAt" sql:"index"`
}

type ViewData struct{
	// Template EmailTemplate
	TemplateName string
	// User User
	User map[string]interface{}
	// Json map[string](string)
	Json map[string]interface{}
}

func (EmailTemplate) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&EmailTemplate{})
	db.Exec("ALTER TABLE email_templates \n ADD CONSTRAINT email_templates_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\n")

}

// ########### CRUD FUNCTIONAL #########

func (et *EmailTemplate) BeforeCreate(scope *gorm.Scope) error {
	et.ID = 0
	et.HashID = strings.ToLower(utils.RandStringBytesMaskImprSrcUnsafe(12, true))
	et.CreatedAt = time.Now().UTC()

	return nil
}

func (et EmailTemplate) create() (*EmailTemplate, error)  {
	err := db.Create(&et).Error
	return &et, err
}

func (EmailTemplate) get(id uint) (*EmailTemplate, error)  {

	et := EmailTemplate{}

	err := db.First(&et, id).Error
	if err != nil {
		return nil, err
	}
	return &et, nil
}

func (EmailTemplate) getByHashId(hashId string) (*EmailTemplate, error) {
	et := EmailTemplate{}

	err := db.First(&et, "hash_id = ?", hashId).Error
	if err != nil {
		return nil, err
	}
	return &et, nil
}

func (et *EmailTemplate) update(input interface{}) error {
	return db.Model(et).Omit("id", "hashId", "created_at", "deleted_at", "updated_at").Update(structs.Map(input)).Error
}

func (et EmailTemplate) Delete () error {
	return db.Model(EmailTemplate{}).Where("id = ?", et.ID).Delete(et).Error
}

// ########### ACCOUNT FUNCTIONAL ###########

func (account Account) CreateEmailTemplate(et EmailTemplate) (*EmailTemplate, error) {
	et.AccountID = account.ID
	return et.create()
}

func (account Account) EmailTemplateUpdate(et *EmailTemplate, input interface{}) error {

	// check account ID
	if et.AccountID != account.ID {
		return errors.New("Шаблон принадлежит другому аккаунту")
	}

	return et.update(input)
}

func (account Account) GetEmailTemplate(id uint) (*EmailTemplate, error) {

	et, err := (EmailTemplate{}).get(id)
	if err != nil {
		return nil, err
	}

	if et.AccountID != account.ID {
		return nil, errors.New("Шаблон принадлежит другому аккаунту")
	}

	return et, nil

}

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

func (account Account) GetEmailTemplates() ([]EmailTemplate, error) {
	var templates []EmailTemplate
	err := db.Find(&templates, "account_id = ?", account.ID).Error
	return templates, err
}

func (account Account) EmailTemplatesList() ([]EmailTemplate, error) {
	
	var templates []EmailTemplate

	// Without Code string
	err := db.Select([]string{"hash_id", "public", "name", "updated_at", "created_at"}).Find(&templates, "account_id = ?", account.ID).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		fmt.Println("Error email templates: ", err)
		return nil, err
	}
	
	return templates, nil
}

// ########### END OF ACCOUNT FUNCTIONAL ###########

// Подготавливает данные для отправки обезличивая их
func (et EmailTemplate) ViewData(user User) (*ViewData, error) {

	// 1. Готовим JSON
	jsonMap := make(map[string]interface{})
	err := et.Json.AssignTo(&jsonMap)
	if err != nil {
		return nil, errors.New("Json data not valid")
	}

	return &ViewData{
		TemplateName: et.Name, // ? надо ли?
		User: *user.DepersonalizedDataMap(),
		Json: jsonMap,
	}, nil
}

// Возвращает тело письма в формате string в кодировке HTML, учитывая переменные в T[map]
// func (et EmailTemplate) GetHTML(T interface{}) (html string, err error) {
func (et EmailTemplate) GetHTML(viewData *ViewData) (html string, err error) {

	body := new(bytes.Buffer)

	tmpl, err := template.New(et.Name).Parse(et.Code)
	if err != nil {
		return "", err
	}

	err = tmpl.Execute(body, viewData)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Ошибка email-шаблона: %s\r", err))
	}
	
	return body.String(), nil
}

// publish Email after execute template
func (account Account) PublishEmail(et EmailTemplate, vData *ViewData) (e *EnvelopePublished, err error) {

	// 1. Проверка (?)
	// todo

	// 2. Результат в виде html сохраняем в аккаунт
	html, err := et.GetHTML(vData)
	if err != nil {
		return nil, err
	}
	
	e, err = account.CreateEnvelopePublishes( EnvelopePublished{HashID: utils.RandStringBytes(5), Body: html} )
	if err != nil {
		return nil, err
	}


	return e, nil
}

// user - получатель письма
// func (et EmailTemplate) Send(from EmailBox, user User, subject string, json map[string](string)) error {
func (et EmailTemplate) Send(from EmailBox, user User, subject string, json map[string]interface{}) error {

	// Принадлежность пользователя к аккаунту не проверяем, т.к. это пофигу
	// user - получатель письма, письмо уйдет на user.Email

	// Формируем данные для сборки шаблона
	// vData := ViewData{et, user, json}
	vData, err := et.ViewData(user)

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

	// privRSAKey := "-----BEGIN RSA PRIVATE KEY-----\nMIICXQIBAAKBgQC4dksLEYhARII4b77fe403uCJhD8x5Rddp9aUJCg1vby7d6QLO\npP7uXpXKVLXxaxQcX7Kjw2kGzlvx7N+d2tToZ8+T3SUadZxLOLYDYkwalkP3vhmA\n3cMuhpRrwOgWzDqSWsDfXgr4w+p1BmNbScpBYCwCrRQ7B12/EXioNcioCQIDAQAB\nAoGAJnnWMVrY1r7zgp4cbDUzQZoQ4boP5oPg6OMqJ3aHUuUYG4WM5lmYK1RjXi7J\nPLAfI8P6WRpbf+XvW8kS47RPkEdXa7svHYa7NT1jQKWY9FwQm1+unc65oK0rZrvE\nrVK0TzK1eQmTxI8OSgFQqShkCZgg45wg9I6iJszkD3loORkCQQDyInM8Un30+2Pq\n2jgH+0Kwa+8x5pEOR4TI5UE4JyzUXVxLuoQNTSMrO2B9Ik6G0Xq7xXFrimMOnLA5\nC/6Ck4ILAkEAwwZl+3I6aZ4rf0n789ktf8zh7UfYhrhQD3uhgSlQ53dMxj0VCBCu\nQQZnWt+MKU/bgEkiHC+aer6iUiJ/H94+uwJBAMZDvTYUmfyiaBNi8eRfMiFBkA+9\nKuOVXj4dsoSnV0bg13VO2VgG5Jg+u2hbUg+EscnVB2U2YJwTYxyjHJiQ7jcCQC2p\n5N0QLO8n8sVWHGFHO6kN3uSBCwjYRR6q8vDcLK5Vt6s/CBqgVTyydCbJ6vaNVTbf\naNYyqzgMRNN4ck2S6xsCQQCoXzfKwz+FfsSAr9WGM/twwCoO/GmDNY5BmwfQuziV\nsYqmmvt6WQ2GxNwcx2VJ/yKIqPU8ABmFPptyPgWXZ4i2\n-----END RSA PRIVATE KEY-----"
	privRSAKey := from.Domain.DKIMPrivateRSAKey

	options := dkim.NewSigOptions()
	options.PrivateKey = []byte(privRSAKey)
	//options.Domain = "rtcrm.ru"
	options.Domain = from.Domain.Hostname
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
	err = client.Mail("huq-nana.abuse.@ratuscrm.com")
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

	// fmt.Println("Сообщение успешно отправлено")

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