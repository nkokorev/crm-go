package models

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/emersion/go-msgauth/dkim"
	"html/template"
	"io"
	"log"
	"net"
	"net/mail"
	"net/smtp"
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

	mailbox := domain.MailBoxes[0]
	if &mailbox == nil {
		return errors.New("Нет mailbox")
	}

	// ввод общих данных
	email.SetSubject("Достало уже!!!")
	email.SetTo("nkokorev@rus-marketing.ru")
	//email.SetTo("mail-test@ratus-dev.ru")
	//email.SetTo("mex388@mail.ru")

	//email.SetReturnPath("abuse@mta1@ratuscrm.com")
	//email.SetMessageID("jd7dhds73h3738")
	//email.SetFrom("Ratus CRM","info@ratuscrm.com")
	email.SetFrom( mailbox.FromName, mailbox.BoxName + "@" + domain.Host)

	err = email.DKIMSign(domain)
	if err != nil {
		return err
	}

	//fmt.Println( domain.GetPrivateKeyByte() )
	//fmt.Println( email.GetMessage() )

	// можем отправлять письмо
	err = email.Send()
	if err != nil {
		return err
	}

	return err
}

// возвращает новое письмо и загружает шаблон при наличии имени файла
func NewEmail(filename string, T interface{}) (*Email, error) {
	email := new(Email)

	// Инициируем хедер
	email.Header = make(map[string]string)
	email.AddHeader("MIME-Version", "1.0")
	email.AddHeader("Content-Type", "text/html; charset=UTF-8")
	//email.AddHeader("Content-Transfer-Encoding", "base64")
	//email.AddHeader("Content-Transfer-Encoding", "binary")
	email.AddHeader("Content-Transfer-Encoding", "7bit")
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

	//body := html.EscapeString(base64.StdEncoding.EncodeToString([]byte(email.GetBody())))
	//body := base64.StdEncoding.EncodeToString([]byte(email.GetBody()))
	/*b64, err := email.GetBodyBase64()
	if err != nil {
		return ""
	}*/
	//return html.EscapeString(email.GetHeader() + "\r\n" + body)
	return email.GetHeader() + "\r\n" + email.GetBody()
}

// подписывает письмо и записывает его в email.BodySignedDKIM
func (email *Email) DKIMSign(domain Domain) error {
	// email is the email to sign (byte slice)
	// privateKey the private key (pem encoded, byte slice )
	/*options := dkim.NewSigOptions()
	options.PrivateKey = domain.GetPrivateKeyByte()
	//options.PrivateKey = []byte(string(domain.DKIMRSAPrivateKey))
	options.Domain = domain.Host
	options.Selector = domain.DKIMSelector
	options.SignatureExpireIn = 3600
	options.BodyLength = uint(len([]rune(email.Body.String()))) // ??
	options.Headers = email.GetHeaders() //[]string{"from", "date", "mime-version", "received", "received"}
	options.AddSignatureTimestamp = false
	options.Canonicalization = "relaxed/relaxed"*/

	r := strings.NewReader(email.GetMessage())

	options := &dkim.SignOptions{
		Domain: domain.Host,
		Selector: domain.DKIMSelector,
		Signer: domain.GetPrivateKey(),
		//HeaderCanonicalization: dkim.CanonicalizationRelaxed,
		//BodyCanonicalization: dkim.CanonicalizationRelaxed,
	}

	var bodyDkim bytes.Buffer
	if err := dkim.Sign(&bodyDkim, r, options); err != nil {
		log.Fatal(err)
	}

	email.BodySignedDKIM = bodyDkim.Bytes()

	return nil
}

var (
	ports = []int{25, 2525, 587}
)

func (email Email) Send() error {
	if !strings.Contains(email.To.Address, "@") {
		return fmt.Errorf("Invalid recipient address: <%s>", email.To.Address)
	}

	host := strings.Split(email.To.Address, "@")[1]
	addrs, err := net.LookupMX(host)
	if err != nil {
		return err
	}

	c, err := newClient(addrs, ports)
	if err != nil {
		return err
	}

	err = email.send(c)
	if err != nil {
		return err
	}

	return nil
}

func newClient(mx []*net.MX, ports []int) (*smtp.Client, error) {
	for i := range mx {
		for j := range ports {
			server := strings.TrimSuffix(mx[i].Host, ".")
			hostPort := fmt.Sprintf("%s:%d", server, ports[j])

			conn, err := net.DialTimeout("tcp", hostPort, 5*time.Second)
			if err != nil {
				if j == len(ports)-1 {
					return nil, err
				}

				continue
			}

			client, err := smtp.NewClient(conn, server)


			//client, err := smtp.Dial(hostPort)
			if err != nil {
				if j == len(ports)-1 {
					return nil, err
				}

				continue
			}
			tlc := &tls.Config{
				InsecureSkipVerify: true,
				ServerName:         server,
			}
			if err := client.StartTLS(tlc); err != nil {
				fmt.Println("Не удалось установить TLC")
			}

			return client, nil
		}
	}

	return nil, fmt.Errorf("Couldn't connect to servers %v on any common port.", mx)
}

func (email Email) send(c *smtp.Client) error {

	//fmt.Println(email.GetMessage() )
	//return nil

	if err := c.Mail(email.From.Address); err != nil {
		log.Println("c.Mail")
		return err
	}

	if err := c.Rcpt(email.To.Address); err != nil {
		log.Println("c.Rcpt")
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	//_, err = fmt.Fprint(w, email.GetMessage())
	_, err = w.Write(email.BodySignedDKIM)
	if err != nil {
		return err
	}

	//defer w.Close()
	err = w.Close()
	if err != nil {
		return err
	}

	//defer c.Quit()

	err = c.Quit()
	if err != nil {
		return err
	}

	return nil
}