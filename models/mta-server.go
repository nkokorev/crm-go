package models

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/toorop/go-dkim"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"sync"
	"time"
)

func init() {
	smtpCh = make(chan EmailPkg, 3)
	go mtaServer(smtpCh) // start MTA server
}

var smtpCh chan EmailPkg

type EmailPkg struct {
	From 		mail.Address
	To 			mail.Address
	Subject     string
	Domain 		Domain // for DKIM

	EmailTemplate EmailTemplate // сам шаблон письма
	ViewData	ViewData
	
	Account       Account
}

type ViewData struct{
	TemplateName string
	PreviewText string
	User map[string]interface{}
	Json map[string]interface{}
}

func mtaServer(c <-chan EmailPkg) {

	var wg sync.WaitGroup
	workerCount := 3

	// add max gorutines
	// wg.Add(workerCount)

	// target speed: 16 mail per second (62 ms / 1 mail)
	for {

		for i := 0; i < workerCount; i++ {

			select {
			 case pkg := <- c:

				 // fmt.Printf("Принял сообщение: %s \n", pkg.Subject)
				 // fmt.Printf("В очереди: %d\n", len(c))
				 // fmt.Printf("Макс. длина: %d\n", cap(c))

				 // Без go - ожидает отправки каждого сообщения
				 wg.Add(1)
				 go mtaSender(pkg, &wg)
			default:
			 	time.Sleep(time.Second*1)
			}

		}

		/*select {
		 case pkg := <- c:
		 	
			 // fmt.Printf("Принял сообщение: %s \n", pkg.Subject)
			 // fmt.Printf("В очереди: %d\n", len(c))
			 // fmt.Printf("Макс. длина: %d\n", cap(c))

			 // Без go - ожидает отправки каждого сообщения
			 wg.Add(1)
			 // workerCount--
			 go mtaSender(pkg, &wg)
		// default:
		// 	time.Sleep(time.Second*1)
		}*/
		wg.Wait()
		// имитируем его отправку
		// time.Sleep(time.Millisecond*100)
	}
}

// Функция по отправке почтового пакета, обычно, работает в отдельной горутине
func mtaSender(pkg EmailPkg, wg *sync.WaitGroup) {
	defer wg.Done()
	time.Sleep(time.Second*2)
	fmt.Println("msg sent...")
	return
	
	// 1. Получаем compile html из email'а
	html, err := pkg.EmailTemplate.GetHTML(&pkg.ViewData)
	if err != nil {
		skipSend(err)
		return
	}

	// 2. Собираем хедеры
	headers := getHeaders(pkg.From, pkg.To, pkg.Subject, "1002", "1324078:20488:trust:54854")

	// 3. Создаем тело сообщения с хедерами и html
	body, err := getSignBody(headers, html, pkg.Domain)
	if err != nil {
		skipSend(err)
		return
	}

	// 4. Делаем коннект к почтовому серверу получателя
	client, err := getClientByEmail(pkg.To.Address)
	if err != nil {
		skipSend(err)
		return
	}

	// 5. Отсылаем сообщение
	err = sendMailByClient(client,body,pkg.To.Address)
	if err != nil {
		skipSend(err)
		return
	}
}

func SendEmailPkg(pkg EmailPkg)  {
	smtpCh <- pkg
}

func skipSend(err error)  {
	fmt.Println("Error: ", err)
}

func getHeaders(from, to mail.Address, subject string, messageId, feedbackId string) *map[string]string {
	if len([]rune(messageId)) > 40 {
		messageId = "101"
	}
	
	headers := make(map[string]string)

	address := from
	headers["From"] = address.String()
	headers["To"] = to.Address
	headers["Subject"] = subject

	// Статичные хедеры
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"
	headers["Content-Transfer-Encoding"] = "quoted-printable"
	headers["Feedback-ID"] = feedbackId //"1324078:20488:trust:54854"
	// Идентификатор представляет собой 32-битное число в диапазоне от 1 до 2147483647, либо строку длиной до 40 символов, состоящую из латинских букв, цифр и символов ".-_".
	//List-Unsubscribe-Post: List-Unsubscribe=One-Click
	//List-Unsubscribe: <https://your-company-net/unsubscribe/example>
	headers["Message-ID"] = messageId // номер сообщения (внутренний номер)
	headers["Received"] = "RatusCRM"  // имя SMTP сервера

	return &headers
}

func getOptionsForDKIM(domain Domain, headers map[string]string) dkim.SigOptions {
	options := dkim.NewSigOptions()
	options.PrivateKey = []byte(domain.DKIMPrivateRSAKey)
	options.Domain = domain.Hostname
	options.Selector = "dk1"
	options.SignatureExpireIn = 0
	options.BodyLength = 50
	options.Headers = GetHeaderKeys(headers)
	options.AddSignatureTimestamp = false
	options.Canonicalization = "relaxed/relaxed"

	return options
}

func getSignBody(headers *map[string]string, html string, domain Domain) ([]byte, error) {

	message := "" // return value

	for k,v := range *headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	var buf bytes.Buffer // body of message
	w := quotedprintable.NewWriter(&buf)
	_, err := w.Write([]byte(html))
	if err != nil {
		return nil, err
	}

	if err = w.Close(); err != nil {
		return nil, err
	}

	message += "\r\n" + buf.String()

	// try DKIM
	dkimOptions := getOptionsForDKIM(domain, *headers)

	body := []byte(message)
	if err := dkim.Sign(&body, dkimOptions); err != nil {
		return nil, errors.New("Cant sign DKIM")
	}

	return body, nil
}

func getHostFromEmail(email string) (account, host string, err error) {
	i := strings.LastIndexByte(email, '@')
	if i == -1 {
		return "", "", errors.New("Email не корректен")
	}
	account = email[:i]
	host = email[i+1:]
	return
}

func getClientByEmail(email string) (*smtp.Client, error) {
	// 4. Получаем хост, на который нужно отправить email
	_, host, err := getHostFromEmail(email)
	if err != nil {
		return nil, err
	}

	mx, err := net.LookupMX(host)
	if err != nil {
		return nil, err
	}

	// список портов, к которым пробуем подключиться
	var ports = []int{25, 2525, 587}
	var client = new(smtp.Client)

	for i := range mx {
		for j := range ports {
			server := strings.TrimSuffix(mx[i].Host, ".")
			hostPort := fmt.Sprintf("%s:%d", mx[i].Host, ports[j])

			// Время в течении которого пытаемся установить связь
			conn, err := net.DialTimeout("tcp", hostPort, 5*time.Second)
			if err != nil {
				fmt.Printf("Коннект не прошел: %s\n", hostPort)
				if j == len(ports)-1 {
					return nil, err
				}

				continue
			}

			// _client, err := smtp.Dial(conn, server)
			_client, err := smtp.NewClient(conn, server)
			if err != nil {
				fmt.Printf("Не удалось подключиться: %s\n", server)
				if j == len(ports)-1 {
					return nil, err
				}
				continue
			}

			tlc := &tls.Config{
				InsecureSkipVerify: true,
				// ServerName: host,
				ServerName: server,
			}
			if err := _client.StartTLS(tlc); err != nil {
				fmt.Println("Не удалось установить TLC")
			}

			client = _client
			break
		}
	}

	return client, nil

}

func sendMailByClient(client *smtp.Client, body []byte, to string) error {

	defer client.Close()
	
	err := client.Mail("userId.abuse.@ratuscrm.com")
	if err != nil {
		return errors.New("Почтовый адрес не может принять почту")
	}

	err = client.Rcpt(to)
	if err != nil {
		return errors.New("Похоже, почтовый адрес не сущесвует")
	}

	wc, err := client.Data()
	if err != nil {
		return errors.New("Клиент не готовы принять сообщение")
	}

	_, err = wc.Write(body)
	if err != nil {
		return errors.New("Не удалось отправить сообщение")
	}

	err = wc.Close()
	if err != nil {
		return errors.New("Ошибка закрытия коннекта")

	}

	// Send the QUIT command and close the connection.
	err = client.Quit()
	if err != nil {
		return errors.New("Ошибка закрытия коннекта 2")
	}


	return nil
}
