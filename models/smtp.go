package models

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/toorop/go-dkim"
	"html/template"
	"io"
	"log"
	"net/mail"
	"strings"
	"time"
)

type Email struct {
	//Message mail.Message // has header and body
	Header map[string]string
	Body   bytes.Buffer
	BodySignedDKIM []byte
	Subject string
	To mail.Address
	From mail.Address

	Tpl *template.Template // шаблон письма с которого читаются данные
}

func TestSend() error {

	// создаем письмо и сразу подгружаем шаблончик
	email, err := NewEmail("example", nil)
	if err != nil { return err }

	// ввод общих данных
	email.SetSubject("Тестовое сообщение")
	email.SetTo("nkokorev@rus-marketing.ru")
	email.SetFrom("Ratus CRM","info@ratuscrm.com")
	email.SetReturnPath("abuse@mta1@ratuscrm.com")
	email.SetMessageID("jd7dhds73h3738")


	// test body64 string
	//body, err := email.GetBodyBase64()
	//fmt.Println(body)

	//header := email.GetHeaderByte()
	//fmt.Println(header.Bytes())

	//fmt.Println(email.GetBodyBase64())

	//fmt.Println(email.GetRSAPrivateKey("-----BEGIN RSA PRIVATE KEY-----\nMIICXQIBAAKBgQCwy7WZIg2haroLTj14GS7MVeLyR0RE7hkhdPYVjdKlUlaJeun5\nlwp7//QcQmZPu9O7e46mTD+CE6srCVyKWCSeUlAVwcV7GT7A9VKnPPiGgAs26Hqz\nAuGwhER3l+lT1arVTbRu7E6shBoWROwPAqZPPp+jctL79CEta5U2ICduHQIDAQAB\nAoGAE+aKRXd400+hK36eGrOy+ds9FYqCG8Q1Xfe9b4WsTWGsTgNg7PBchMK15qxu\nudDpr3PkBcIVb/3oyYpfOU9cp6mgXk557OxqfPNyNwRO/o/6/IiEpFFrk8jJxoc3\nmoa9Lh1hM/lsSGryp83L1vBUTs3tXIGo+uBBHnLaH33dFF0CQQDhizg/xVAhR4he\n8Q/uSP5Cgf/Viwevluxpz2R4WrGro5XRyLvEoXb+gPG9NqjT62N7jHX1lBxFpFPT\n/zh1BADLAkEAyKtTmww6/ULKTijfBOhp+w/O4TOWbq0JSZBXAGPI6jh+73gGNf/x\n+55kMYUjIaxpIkILsDTlQrO5kBIBarX3twJAHtXp2s4fJm2hN1m909Ym7PDZCVj4\ntAjuSYkRM2My50R2Nzg6c6efnSwD4NqYOmD0OO/7MJgPRXYx/8nk7hqeAQJBAJ96\n8h42cSdYjpnhh6VJ5PigTqXSLwtUwB3T9iEcLNBhCBjfhegiurlj33MvwYUAlimg\n3dMzpsUFO0PR24hoiC8CQQDP1kDw2zzA8dwGFjbBPqFfN5uVcbwzq1tRjdM1mkp8\nwJB/anwuIRNIE/PDCvi4MEmW7p7FkfbHOZOSgYXbIK3k\n-----END RSA PRIVATE KEY-----").D)

	acc, err := GetMainAccount()
	if err != nil { return err }
	domains, err := acc.GetDomains()
	if err != nil || domains == nil {
		return errors.New("Не удалось получить домены для аккаунта")
	}
	if len(domains) == 0 {
		return errors.New("У аккаунта нет доступных доменов для отправки писем")
	}

	// получаем рабочий домен с которого будем отсылать сообщения
	domain := domains[0]

	err = email.DKIMSign(domain)
	if err != nil {
		return err
	}

	//fmt.Println( domain.GetPrivateKeyByte() )
	fmt.Println( email.BodySignedDKIM )


	return err
}

// возвращает новое письмо и загружает шаблон при наличии имени файла
func NewEmail(filename string, T interface{}) (*Email, error) {
	email := new(Email)

	// Инициируем хедер
	email.Header = make(map[string]string)
	email.AddHeader("MIME-Version", "1.0")
	email.AddHeader("Content-Transfer-Encoding", "base64")
	email.AddHeader("Content-Type", "text/html; charset=utf-8")
	email.AddHeader("Date", time.RFC1123Z)

	if filename != "" {
		err := email.LoadBodyTemplate(filename + ".html", T)
		if err != nil {
			return nil, err
		}

	}

	return email, nil
}

func (email *Email) SetSubject(v string) {
	email.Subject = v
	email.AddHeader("Subject", email.Subject)
}

func (email *Email) SetTo(addr string) {
	email.To = mail.Address{Address: addr}
	email.AddHeader("To", email.To.Address)
}
func (email *Email) SetFrom(name, addr string) {
	if name != "" {
		email.From = mail.Address{Name: name, Address: addr}
	} else {
		email.From = mail.Address{Address: addr}
	}

	email.AddHeader("From", email.From.String())
}
func (email *Email) SetReturnPath(path string) {
	email.AddHeader("Return-Path", path)
}
func (email *Email) SetMessageID(id string) {
	email.AddHeader("Message-ID", id)
}

// Загружает шаблон письма (html.template)
func (email *Email) LoadBodyTemplate(filename string, T interface{}) (err error) {

	// тут можно сделать какой-то поиск среди доступных шаблонов
	email.Tpl, err = template.ParseFiles("files/" + filename)
	if err != nil {
		return err
	}

	if err = email.ExecuteBodyTemplate(T); err != nil {
		return err
	}

	return nil
}

// подгружает данные в шаблон и результат в тело письма
func (email *Email) ExecuteBodyTemplate(T interface{}) error {
	var buf = new(bytes.Buffer)

	err := email.Tpl.Execute(buf, T)
	if err != nil {
		return err
	}

	email.Body = *buf

	return nil
}

func (email Email) GetBodyByte() []byte {
	return email.Body.Bytes()
}

// Возвращает Body
func (email Email) GetBody() string {

	return string(email.Body.Bytes()[:])
}

// Возвращает экранированное Body
func (email Email) GetBodyEscaped() string {
	return template.HTMLEscapeString(email.GetBody())
}

type linesplitter struct {
	len   int
	count int
	sep   []byte
	w     io.Writer
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func (ls *linesplitter) Close() (err error) {
	return nil
}
func (ls *linesplitter) Write(in []byte) (n int, err error) {
	writtenThisCall := 0
	readPos := 0
	// Leading chunk size is limited by: how much input there is; defined split length; and
	// any residual from last time
	chunkSize := min(len(in), ls.len-ls.count)
	// Pass on chunk(s)
	for {
		ls.w.Write(in[readPos:(readPos + chunkSize)])
		readPos += chunkSize // Skip forward ready for next chunk
		ls.count += chunkSize
		writtenThisCall += chunkSize

		// if we have completed a chunk, emit a separator
		if ls.count >= ls.len {
			ls.w.Write(ls.sep)
			writtenThisCall += len(ls.sep)
			ls.count = 0
		}
		inToGo := len(in) - readPos
		if inToGo <= 0 {
			break // reached end of input data
		}
		// Determine size of the NEXT chunk
		chunkSize = min(inToGo, ls.len)
	}
	return writtenThisCall, nil
}

// возвращает body в base64 по 76 символов в длину в виде []Byte
func (email Email) GetBodyBase64Byte() ([]byte, error) {

	bufR := new(bytes.Buffer)

	lsWriter := &linesplitter{len: 76, count: 0, sep: []byte("\n"), w: bufR}
	wrt := base64.NewEncoder(base64.StdEncoding, lsWriter)
	_, err := io.Copy(wrt, strings.NewReader(email.GetBodyEscaped()))
	if err != nil {
		log.Fatal(err)
	}

	return bufR.Bytes(), nil
}

// возвращает body в base64 по 76 символов в длину в виде строки
func (email Email) GetBodyBase64() (string, error) {

	bodyByte, err := email.GetBodyBase64Byte()
	if err != nil { return "", err}

	return string(bodyByte[:]), nil

}

func (email Email) GetHeader() string {
	header := ""
		for k, v := range email.Header {
			header += k + ": " + v + "\r\n"
		}
	return header
}

func (email Email) GetHeaderByte() bytes.Buffer {
	buf := new(bytes.Buffer)
	io.Copy(buf, strings.NewReader(email.GetHeader()))

	return *buf
}

func (email *Email) AddHeader(k string, v string) {
	email.Header[k] = v
}

// Возвращает список заголовоков
func (email Email) GetHeaders() []string {

	var headers []string
	for k,_ := range email.Header {
		headers = append(headers, k)
	}

	return headers
}

// Возвращает все сообщение
func (email Email) GetMessage() string {
	return email.GetHeader() + "\r\n" + email.GetBody()
}

// подписывает письмо
func (email *Email) DKIMSign(domain Domain) error {
	// email is the email to sign (byte slice)
	// privateKey the private key (pem encoded, byte slice )
	options := dkim.NewSigOptions()
	options.PrivateKey = domain.GetPrivateKeyByte()
	//options.PrivateKey = []byte(string(domain.DKIMRSAPrivateKey))
	options.Domain = domain.Host
	options.Selector = domain.DKIMSelector
	options.SignatureExpireIn = 3600
	options.BodyLength = 0 //uint(len([]rune(email.Body.String()))) // ??
	options.Headers = email.GetHeaders() //[]string{"from", "date", "mime-version", "received", "received"}
	options.AddSignatureTimestamp = true
	options.Canonicalization = "relaxed/relaxed"

	// Получаем все письма с заголовками
	body, err := email.GetBodyBase64Byte()
	if err != nil {
		return err
	}
	err = dkim.Sign(&body, options)
	if err != nil {
		return err
	}

	email.BodySignedDKIM = body

	return nil
}