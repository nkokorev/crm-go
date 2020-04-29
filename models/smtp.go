package models

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/mail"
	"strings"
	"time"
)

type Email struct {
	//Message mail.Message // has header and body
	Header map[string]string
	Body   io.Reader
	Subject string
	To mail.Address
	From mail.Address
}

func TestSend() error {

	email, err := NewEmail("example", nil)
	if err != nil { return err }

	email.SetSubject("Тестовое сообщение")
	email.SetTo("nkokorev@rus-marketing.ru")
	email.SetFrom("Ratus CRM","info@ratuscrm.com")
	email.SetReturnPath("abuse@mta1@ratuscrm.com")
	email.SetMessageID("jd7dhds73h3738")


	// test body64 string
	//body, err := email.GetBodyBase64String()
	//fmt.Println(body)

	header := email.GetHeaderByte()
	fmt.Println(header.Bytes())

	return err
}

// возвращает новое письмо и загружает шаблон
func NewEmail(filename string, T interface{}) (*Email, error) {
	email := new(Email)

	// Инициируем хедер
	email.Header = make(map[string]string)
	email.AddHeader("MIME-Version", "1.0")
	email.AddHeader("Content-Transfer-Encoding", "base64")
	email.AddHeader("Content-Type", "text/html; charset=utf-8")
	email.AddHeader("Date", time.RFC1123Z)

	err := email.LoadBodyFromTemplate(filename + ".html", T)
	if err != nil {
		return nil, err
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

func (email *Email) LoadBodyFromTemplate(filename string, T interface{}) error {

	tpl, err := template.ParseFiles("files/" + filename)
	if err != nil {
		return err
	}

	var buf = new(bytes.Buffer)
	err = tpl.Execute(buf, T)

	email.Body = buf

	return nil
}

func (email Email) GetBodyByte() ([]byte, error) {
	return ioutil.ReadAll(email.Body)
}

// Возвращает Body
func (email Email) GetBodyString() (string, error) {
	body, err := ioutil.ReadAll(email.Body)
	if err != nil { return "", err}

	return string(body[:]), nil
}

// Возвращает экранированное Body
func (email Email) GetBodyStringEscaped() (string, error) {
	body, err := email.GetBodyString()
	if err != nil {
		return "",err
	}
	return template.HTMLEscapeString(body), nil
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
	body, err := email.GetBodyStringEscaped()
	if err != nil {return nil, err}
	_, err = io.Copy(wrt, strings.NewReader(body))
	if err != nil {
		log.Fatal(err)
	}

	return bufR.Bytes(), nil

}

// возвращает body в base64 по 76 символов в длину в виде строки
func (email Email) GetBodyBase64String() (string, error) {

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

