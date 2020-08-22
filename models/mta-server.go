package models

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	u "github.com/nkokorev/crm-go/utils"
	"github.com/toorop/go-dkim"
	"log"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var smtpCh chan EmailPkg
const deepMTACh = 10 // Объем буфера из пакетов и макс. число горутин по их отправке
const dialTimeout = time.Second * 5 // максимальное время для установки соединения с smtp сервером получателя

// число одновременных запущенных потоков отправки для email. Если один висит, два других ждут.
// Очередь запинается, когда один поток подвисает, а другие потоки разобраны.
// var workerCount = 10

func init() {
	smtpCh = make(chan EmailPkg, deepMTACh)
	go mtaServer(smtpCh) // start MTA server
}

// Содержание письма компилируется из ViewData
type ViewData struct {
	Subject string
	PreviewText string

	Data map[string]interface{}
	Json map[string]interface{}

	UnsubscribeURL string
	PixelURL string // ссылка для пикселя
	PixelHTML string // html <<
}

// worker MTA-Server
func mtaServer(c <-chan EmailPkg) {

	// можно доработать синхронизацию, но после тестов по отправке
	var wg sync.WaitGroup
	var m sync.Mutex

	if c == nil {
		log.Println("Ошибка mtaServer: channel is null!")
		return
	}

	// target speed: 16 mail per second (62 ms / 1 mail) [~4800 / мин]
	for {
		select {
		case pkg := <- c:

			// fmt.Printf("Принял сообщение: %s \n", pkg.Subject)

			// Если все потоки разобраны (из пула = {workerCount}) - ожидаем завершения всех текущих отправок (макс 5с), чтобы не плодить овер коннектов
			// if workerCount < 1 {wg.Wait()}


			// обновляем счетчик WaitGroup
			wg.Add(1)
			
			// -1 рабочий поток для отправки
			// workerCount--

			// Без go - ожидает отправки каждого сообщения
			go mtaSender(pkg, &wg, &m)

			wg.Wait()

			// fix ХЗ чего
			// time.Sleep(time.Millisecond*10)
			
		case <- time.After(1 * time.Second):
			// fmt.Println("Подождали 1 секунду -)")
		/*default:
			time.Sleep(time.Millisecond*500)*/
		}
	}
}

// Функция по отправке почтового пакета, обычно, запускается воркером в отдельной горутине
func mtaSender(pkg EmailPkg, wg *sync.WaitGroup, m *sync.Mutex) {

	m.Lock()
	defer m.Unlock()
	defer wg.Done() // отписываемся о закрытии текущей горутины


	fmt.Println("Типа отправляем...")
	time.Sleep(time.Second*3)
	return

	// defer func() {workerCount++}() // освобождаем счетчик потоков (горутин) по отправке

	// 1. Получаем переменные для отправки письма
	account, err := GetAccount(pkg.accountId)
	if err != nil {
		pkg.stopEmailSender(fmt.Sprintf("Ошибка получения аккаунта [id = %v] при отправки email-сообщения: %v", pkg.accountId, err.Error()))
		return
	}

	user, err := account.GetUser(pkg.userId); if err != nil {
		pkg.stopEmailSender(fmt.Sprintf("Ошибка получения пользователя [id = %v] при отправки email-сообщения: %v", pkg.userId, err.Error()))
		return
	}
	historyHashId := strings.ToLower(u.RandStringBytesMaskImprSrcUnsafe(12, true))

	unsubscribeUrl := account.GetUnsubscribeUrl(user.HashId, historyHashId)
	pixelURL := account.GetPixelUrl(historyHashId)
	hashAddress := mail.Address{Address: historyHashId + "@mta1.ratuscrm.com"}

	// Добавляем во данные письма url отписки и счетчика открытий
	(*pkg.viewData).UnsubscribeURL = unsubscribeUrl
	(*pkg.viewData).PixelURL = pixelURL

	// returnPath := "abuse@mta1.ratuscrm.com"  // тут тоже надо hash@ адресс
	returnPath := "smtp@rus-marketing.ru"  // тут тоже надо hash@ адресс
	messageId :=  hashAddress.Address

	// Готовим фидбэк по смыслу: accountId | userId | ownerId | ownerType | MTA server (= ничего не значит)
	feedBackId := strconv.Itoa(int(pkg.accountId)) + ":" + strconv.Itoa(int(pkg.userId)) + ":" + strconv.Itoa(int(pkg.emailSender.GetId())) + ":" +
		u.ToCamel(pkg.emailSender.GetType()) + ":1"

	// 1. Получаем compile html из email'а
	html, err := pkg.emailTemplate.GetHTML(pkg.viewData)
	if err != nil {
		pkg.stopEmailSender(fmt.Sprintf("Ошибка в синтаксисе email-шаблона id = %v: %v", pkg.emailTemplate.Id, err.Error()))
		return
	}
	
	// 2. Собираем хедеры
	headers := getHeaders (
		mail.Address{Name:pkg.emailBox.Name, Address: pkg.emailBox.Box + "@" + pkg.webSite.Hostname},
		pkg.To, pkg.subject, messageId, feedBackId, unsubscribeUrl, historyHashId)

	// 3. Создаем тело сообщения с хедерами и html
	if pkg.webSite.Id < 1 || pkg.emailBox.Id < 1 {
		pkg.stopEmailSender("Техническая ошибка: не удалось установить отправителя: webSite || emailBox id < 1")
		return
	}

	body, err := getSignBody(headers, html, pkg.webSite)
	if err != nil {
		pkg.bounced(softBounced, fmt.Sprintf("Ошибка в процессе DKIM-подписи письма: %v", err.Error()))
		log.Printf("Ошибка в синтаксисе email-шаблона id = %v: %v", pkg.emailTemplate.Id, err.Error())
		return
	}

	// 4. Делаем коннект к почтовому серверу получателя
	client, bounceLevel, err := getClientByEmail(pkg.To.Address)
	if err != nil {
		pkg.bounced(bounceLevel, fmt.Sprintf("Неудается установить connect с MX-сервером: %v", err.Error()))
		return
	}

	// 5. Отсылаем сообщение | требует времени !
	bounceLevel, err = sendMailByClient(client, body, pkg.To.Address, returnPath)
	if err != nil {
		pkg.bounced(bounceLevel, fmt.Sprintf("Ошибка во время отправки письма: %v", err.Error()))
		return
	}

	// 6. Заносим в историю

	// Обновляем / удаляем задачу в воркере отправки писем для автоматической отправки писем
	queueCompleted := pkg.handleQueue()
	
	history := &MTAHistory{
		HashId:  historyHashId,
		AccountId: pkg.accountId,
		UserId: &pkg.userId,
		Email: user.Email,
		OwnerId: pkg.emailSender.GetId(),
		OwnerType: pkg.emailSender.GetType(),
		EmailTemplateId: &pkg.emailTemplate.Id,
		QueueStepId: &pkg.queueStepId, // выполненный шаг
		QueueCompleted: queueCompleted,
	}

	// отлично, создаем запись в истории!
	_, err = history.create()
	if err != nil {
		log.Printf("Ошибка создания записи в истории отправки email-писем. AccountId [id = %v], OwnerId: [id = %v]: %v", pkg.accountId, pkg.emailSender.GetId(), err.Error())
	} else {
		// fmt.Println("Запись в истории создана!")
	}

}

func SendEmail(pkg EmailPkg)  {
	/*select {
	case smtpCh <- pkg:
		fmt.Println("add msg to message channel")
	}*/
	smtpCh <- pkg
}

func getHeaders(from, to mail.Address, subject string, messageId, feedbackId, unsubscribeUrl, historyHashId string) *map[string]string {
	if len([]rune(messageId)) > 40 {
		messageId = "10001@mta1.ratuscrm.com"
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
	headers["List-Unsubscribe"] = unsubscribeUrl
	headers["List-Unsubscribe-Post"] = "List-Unsubscribe=One-Click"

	//List-Unsubscribe-Post: List-Unsubscribe=One-Click
	//List-Unsubscribe: <https://your-company-net/unsubscribe/example>
	headers["Message-ID"] = messageId // номер сообщения (внутренний номер)
	headers["Received"] = "by MTA of RatusCRM"  // имя SMTP сервера
	headers["X-Mailer"] = "RatusCRM Mailer v2.0.1"  // какая программа

	campaignId := "ratuscrm" + historyHashId
	headers["X-Campaign"] = campaignId  // какая программа
	headers["X-campaignid"] = campaignId  // какая программа


	return &headers
}

func getOptionsForDKIM(domain *WebSite, headers map[string]string) dkim.SigOptions {
	options := dkim.NewSigOptions()
	options.PrivateKey = []byte(domain.DKIMPrivateRSAKey)
	options.Domain = domain.Hostname
	options.Selector = domain.DKIMSelector // "dk1"
	options.SignatureExpireIn = 0
	options.BodyLength = 50
	options.Headers = GetHeaderKeys(headers)
	options.AddSignatureTimestamp = false
	options.Canonicalization = "relaxed/relaxed"

	return options
}

func getSignBody(headers *map[string]string, html string, webSite *WebSite) ([]byte, error) {

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
	dkimOptions := getOptionsForDKIM(webSite, *headers)

	body := []byte(message)
	if err := dkim.Sign(&body, dkimOptions); err != nil {
		return nil, err
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

func getClientByEmail(email string) (*smtp.Client, BounceType, error) {

	// 1. Получаем хост, на который нужно отправить email
	_, host, err := getHostFromEmail(email)
	if err != nil {
		return nil, hardBounced, err
	}

	// 2. Получаем список mx-ов, где может быть почтовый сервер
	mx, err := net.LookupMX(host)
	if err != nil {
		return nil, softBounced, err
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
				log.Printf("Коннект не прошел: %s\n", hostPort)
				if j == len(ports)-1 {
					return nil, softBounced, err
				}
				continue
			}

			// _client, err := smtp.Dial(conn, server)
			_client, err := smtp.NewClient(conn, server)
			if err != nil {
				log.Printf("Не удалось подключиться: %s\n", server)
				if j == len(ports)-1 {
					return nil, softBounced, err
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
				log.Println("Не удалось установить TLC")
			}

			client = _client
			break
		}
	}

	return client, "", nil
}

func sendMailByClient(client *smtp.Client, body []byte, to string, returnPath string) (BounceType, error) {

	defer client.Close()

	err := client.Mail(returnPath)
	if err != nil {
		return hardBounced, err
	}

	err = client.Rcpt(to)
	if err != nil {
		// return errors.New("Похоже, почтовый адрес не сущесвует")
		return hardBounced, err
	}

	wc, err := client.Data()
	if err != nil {
		// return errors.New("Клиент не готовы принять сообщение")
		return hardBounced, err
	}
	defer wc.Close()

	_, err = wc.Write(body)
	if err != nil {
		return softBounced, err
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

	return "", nil
}

func MTALenChannel() uint {
   return uint(len(smtpCh))
}
func MTACapChannel() uint {
	return uint(cap(smtpCh))
}
