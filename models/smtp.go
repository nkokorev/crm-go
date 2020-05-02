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

	Header map[string]string
	Body   bytes.Buffer
	BodySignedDKIM []byte
	Subject string
	To mail.Address
	From mail.Address

	Tpl *template.Template // шаблон письма с которого читаются данные
}

func SendTestMessage() error {

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
	email.SetSubject("From local server PEM Certs V")
	email.SetTo("test-d8hbwwoak@srv1.mail-tester.com")
	//email.SetTo("nkokorev@rus-marketing.ru")
	//email.SetTo("mail-test@ratus-dev.ru")
	//email.SetTo("mex388@mail.ru")
	email.SetReturnPath("abuse@mta1@ratuscrm.com")
	email.SetMessageID("jd7dhds73h3738")

	email.SetFrom( mailbox.FromName, mailbox.BoxName + "@" + domain.Host)

	err = email.DKIMSign(domain)
	if err != nil {
		return err
	}

	// можем отправлять письмо
	err = email.Send()
	if err != nil {
		return err
	}

	return err
}

func (email *Email) GetAuthSMTP() smtp.Auth {
	return smtp.PlainAuth("", "mex388", "", "mta1.ratuscrm.com")
}

// возвращает новое письмо и загружает шаблон при наличии имени файла
func NewEmail(filename string, T interface{}) (*Email, error) {
	email := new(Email)

	// Инициируем хедер
	email.Header = make(map[string]string)
	email.AddHeader("MIME-Version", "1.0")
	email.AddHeader("Content-Type", "text/html; charset=UTF-8")
	email.AddHeader("Content-Transfer-Encoding", "base64")
	//email.AddHeader("Content-Transfer-Encoding", "binary")
	//email.AddHeader("Content-Transfer-Encoding", "7bit")
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

func (email Email) GetBody() []byte {
	return email.Body.Bytes()
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
		_, err := ls.w.Write(in[readPos:(readPos + chunkSize)])
		if err != nil {
			return 0, err
		}
		readPos += chunkSize // Skip forward ready for next chunk
		ls.count += chunkSize
		writtenThisCall += chunkSize

		// if we have completed a chunk, emit a separator
		if ls.count >= ls.len {
			_, err := ls.w.Write(ls.sep)
			if err != nil {
				return 0, err
			}
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
func (email Email) GetBodyBase64() ([]byte, error) {

	bufR := new(bytes.Buffer)

	lsWriter := &linesplitter{len: 76, count: 0, sep: []byte("\n"), w: bufR}
	wrt := base64.NewEncoder(base64.StdEncoding, lsWriter)
	//_, err := io.Copy(wrt, strings.NewReader(email.GetBody()))
	/*if err != nil {
		log.Fatal(err)
	}*/

	_, err := wrt.Write(email.GetBody())
	if err != nil {
		return nil, err
	}

	return bufR.Bytes(), nil
}

func (email Email) GetHeader() string {
	header := ""
		for k, v := range email.Header {
			header += k + ": " + v + "\r\n"
		}
	return header
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
func (email Email) GetMessage() ([]byte, error) {

	//body := html.EscapeString(base64.StdEncoding.EncodeToString([]byte(email.GetBody())))
	//body := base64.StdEncoding.EncodeToString([]byte(email.GetBody()))
	/*b64, err := email.GetBodyBase64()
	if err != nil {
		return ""
	}*/
	//return html.EscapeString(email.GetHeader() + "\r\n" + body)

	body, err := email.GetBodyBase64()
	//body := email.GetBody()
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	_, err = buf.Write([]byte(email.GetHeader()+ "\r\n"))
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// подписывает письмо и записывает его в email.BodySignedDKIM
func (email *Email) DKIMSign(domain Domain) error {

	body, err := email.GetMessage()
	if err != nil {
		return err
	}

	//r := strings.NewReader(body)

	options := &dkim.SignOptions{
		Domain: domain.Host,
		Selector: domain.DKIMSelector,
		Signer: domain.GetPrivateKey(),
		HeaderCanonicalization: dkim.CanonicalizationRelaxed,
		BodyCanonicalization: dkim.CanonicalizationRelaxed,
	}

	var bodyDkim bytes.Buffer
	//if err := dkim.Sign(&bodyDkim, r, options); err != nil {
	if err := dkim.Sign(&bodyDkim, bytes.NewBuffer(body), options); err != nil {
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

	client, err := newClient(addrs, ports)
	if err != nil {
		return err
	}
	// получаем авторизацию на SMTP-сервере
	//auth := email.GetAuthSMTP()

	//err := email.send("mta1.ratuscrm.com:25", auth)
	err = email.send(client)
	if err != nil {
		return err
	}

	return nil
}

func (email *Email) GetHostPort(mx []*net.MX, ports []int) (hostPort string, err error) {
	for i := range mx {
		for j := range ports {
			server := strings.TrimSuffix(mx[i].Host, ".")
			hostPort = fmt.Sprintf("%s:%d", server, ports[j])


			_, err := net.DialTimeout("tcp", hostPort, 10*time.Second)
			if err != nil {
				if j == len(ports)-1 {
					return "", err
				}
				continue
			}
		}
	}


	fmt.Println("Найден хост: ", hostPort)

	return hostPort, nil
}

func (email Email) send(c *smtp.Client) error {

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
	//_, err = fmt.Fprint(w, email.GetMessageDKIMSign())
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