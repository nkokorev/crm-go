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

var smtpCh chan EmailPkg
const deepMTACh = 10000 // глубина очереди из пакетов сообщений
const dialTimeout = time.Second * 5 // время для установки коннекта с smtp сервером получателя

// число одновременных запущенных потоков отправки для email. Если один висит, два других ждут.
// Очередь запинается, когда один поток подвисает, а другие потоки разобраны.
var workerCount = 3

func init() {
	smtpCh = make(chan EmailPkg, deepMTACh)
	go mtaServer(smtpCh) // start MTA server
}

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

	var wg sync.WaitGroup // можно доработать синхронизацию, но после тестов по отправке

	// target speed: 16 mail per second (62 ms / 1 mail)
	for {
		select {
		case pkg := <- c:

			// fmt.Printf("Принял сообщение: %s \n", pkg.Subject)
			// fmt.Printf("В очереди: %d\n", len(c))
			// fmt.Printf("Макс. длина: %d\n", cap(c))

			// Если все потоки разобраны (из пула = {workerCount}) - ожидаем завершения всех текущих отправок (макс 5с), чтобы не плодить овер коннектов
			if workerCount < 1 {
				wg.Wait()
			}

			// обновляем счетчик WaitGroup
			wg.Add(1)
			// -1 рабочий поток для отправки
			workerCount--

			// Без go - ожидает отправки каждого сообщения
			go mtaSender(pkg, &wg)
		default:
			time.Sleep(time.Millisecond*100)
		}
	}
	/*for {
		for i := 0; i < workerCount; i++ {
			select {
			 case pkg := <- c:

				 // fmt.Printf("Принял сообщение: %s \n", pkg.Subject)
				 // fmt.Printf("В очереди: %d\n", len(c))
				 // fmt.Printf("Макс. длина: %d\n", cap(c))

				 // обновляем счетчик WaitGroup
				 if workerCount < 1 {
				 	wg.Wait()
				 }
				 wg.Add(1)
				 workerCount--
				 
				 // Без go - ожидает отправки каждого сообщения
				 go mtaSender(pkg, &wg)
			default:
			 	time.Sleep(time.Millisecond*100)
			}
		}

		// ждем завершения предыдущих отправок в {workerCount} потоков
		// wg.Wait()
	}*/
}

// Функция по отправке почтового пакета, обычно, работает в отдельной горутине
func mtaSender(pkg EmailPkg, wg *sync.WaitGroup) {

	defer wg.Done() // отписываемся о закрытии текущей горутины
	defer func() {workerCount++}() // освобождаем счетчик потоков (горутин) по отправке

	// todo: сделать осознанные messageID feedbackId
	// todo: сделать осознанный returnPath
	returnPath := "abuse@ratuscrm.com"
	messageId := "1002"
	feedBackId := "1324078:20488:trust:54854"


	// 1. Получаем compile html из email'а
	html, err := pkg.EmailTemplate.GetHTML(&pkg.ViewData)
	if err != nil {
		skipSend(err)
		return
	}

	// 2. Собираем хедеры
	headers := getHeaders(pkg.From, pkg.To, pkg.Subject, messageId, feedBackId)

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

	// 5. Отсылаем сообщение | требует времени !
	err = sendMailByClient(client, body, pkg.To.Address, returnPath)
	if err != nil {
		skipSend(err)
		return
	}
}

func SendEmail(pkg EmailPkg)  {
	select {
	case smtpCh <- pkg:
		fmt.Println("add msg to message channel")
	}
	// smtpCh <- pkg
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

	// 1. Получаем хост, на который нужно отправить email
	_, host, err := getHostFromEmail(email)
	if err != nil {
		return nil, err
	}

	// 2. Получаем список mx-ов, где может быть почтовый сервер
	mx, err := net.LookupMX(host)
	if err != nil {
		return nil, err
	}
	
	// список портов, к которым пробуем подключиться
	var ports = []int{25, 2525, 587}
	var client = new(smtp.Client)

	// вообще-то надо подключаться к mx с наивысшем рейтингом.. ну да ладно =)
	for i := range mx {
		for j := range ports {
			server := strings.TrimSuffix(mx[i].Host, ".")
			hostPort := fmt.Sprintf("%s:%d", mx[i].Host, ports[j])

			// Ждем {dialTimeout} секунд, для установки связи..
			conn, err := net.DialTimeout("tcp", hostPort, dialTimeout)
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
				ServerName: mx[i].Host, // куда реально будем коннектиться
				// ServerName: host,
				// ServerName: server,
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

func sendMailByClient(client *smtp.Client, body []byte, to string, returnPath string) error {

	defer client.Close()

	err := client.Mail(returnPath)
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
	defer wc.Close()

	_, err = wc.Write(body)
	if err != nil {
		return errors.New("Не удалось отправить сообщение")
	}

	// может вернуть (yandex): 250 2.0.0 Ok: queued on mxfront9q.mail.yandex.net as 1590262878-g7S0CPwaoz-fIVS8Xqq
	// может вернуть (gmail): 250 2.0.0 OK  1590263113 j9si5638746ots.198 - gsmtp
	// err = возвращает фактический статус, без него сообщение не отправляется
	err = client.Quit()
	if err != nil {
		fmt.Println(err)
		// todo: тут нуежн парсер результата
		// return errors.New("Ошибка закрытия коннекта 1")
	}

	// tls: use of closed connection
	// позволяет закрыть коннект в случае успешной отправки письма
	/*err = client.Close()
	if err != nil {
		fmt.Println(err)
		return errors.New("Ошибка закрытия коннекта 2")
	}*/

	return nil
}
