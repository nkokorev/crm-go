package models

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/emersion/go-msgauth/dkim"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

var (
	ports = []int{25, 2525, 587}
)

type Message struct {

	To      mail.Address
	From    mail.Address
	Subject string
	Body    string
}

// Send sends a message to recipient(s) listed in the 'To' field of a Message
func (m Message) Send() error {
	if !strings.Contains(m.To.Address, "@") {
		return fmt.Errorf("Invalid recipient address: <%s>", m.To)
	}

	host := strings.Split(m.To.Address, "@")[1]
	addrs, err := net.LookupMX(host)
	if err != nil {
		return err
	}

	c, err := newClient(addrs, ports)
	if err != nil {
		return err
	}

	err = send(m, c)
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

func send(m Message, c *smtp.Client) error {

	if err := c.Mail(m.From.Address); err != nil {
		log.Println("c.Mail")
		return err
	}

	if err := c.Rcpt(m.To.Address); err != nil {
		log.Println("c.Rcpt")
		return err
	}

	header := make(map[string]string)
	header["Return-Path"] = "<bounce@ratuscrm.com>"
	//header["Precedence"] = "bulk"
	header["Feedback-ID"] = "vtvent-15:nkokorev@rus-marketing.ru:campaign:RatusSMTP"
	header["From"] = m.From.String()
	header["To"] = m.To.String()
	header["Subject"] = m.Subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=utf-8"
	header["Content-Transfer-Encoding"] = "base64"
	header["Return-Path"] = "<bounce@ratuscrm.com>"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(m.Body))

	mTpl := signDKIM(message)
	message = mTpl.String()

	msg, err := c.Data()
	if err != nil {
		return err
	}

	/*if m.Subject != "" {
		_, err = msg.Write([]byte("Subject: " + m.Subject + "\r\n"))
		if err != nil {
			return err
		}
	}

	if m.From != "" {
		_, err = msg.Write([]byte("From: <" + m.From + ">\r\n"))
		if err != nil {
			return err
		}
	}

	if m.To != "" {
		_, err = msg.Write([]byte("To: <" + m.To + ">\r\n"))
		if err != nil {
			return err
		}
	}

	_, err = msg.Write([]byte("Subject: " + m.Subject + "\r\n"))
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(msg, m.Body)
	if err != nil {
		return err
	}
	*/

	_, err = fmt.Fprint(msg, message)
	if err != nil {
		return err
	}

	err = msg.Close()
	if err != nil {
		return err
	}

	err = c.Quit()
	if err != nil {
		return err
	}

	return nil
}

func signDKIM (mailString string) bytes.Buffer {
	r := strings.NewReader(mailString)

	//rsaPrivateKey := "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDSw3hDW4hWLBZ2tLEZhl6lkXuBkxngTKoJm7Rim5pOGPuoxMekfKhj1egQA8Kh/+FnKVXJP/fsQpmoCGjxCdjC8dhUzUbZIj8OhBnMsa3uaAyOHNXnBWnZVfXSjtOQVfpJltt+SHy/CptXuX7TyvXZt65OdmKjvfHyvsByJEqdUwIDAQAB"
	//rsaPrivateKey := "-----BEGIN RSA PRIVATE KEY-----\nMIICXAIBAAKBgQDSw3hDW4hWLBZ2tLEZhl6lkXuBkxngTKoJm7Rim5pOGPuoxMek\nfKhj1egQA8Kh/+FnKVXJP/fsQpmoCGjxCdjC8dhUzUbZIj8OhBnMsa3uaAyOHNXn\nBWnZVfXSjtOQVfpJltt+SHy/CptXuX7TyvXZt65OdmKjvfHyvsByJEqdUwIDAQAB\nAoGAIDA8OMVM8CQxlhWIiqZr5Atw+lwV8pyix27hQMIU8eJ85MyQ1P041m5/z5pT\nalxi91dnw6GiYpHVV8VZCZ8AXJaP4f8DFOC5oAI6qe7xHtilnI3p8cItRVTEi6uU\nEamyh5fwZcVzq8yFAvIWc4P1bC5xJ8ce3P2QPVhpojIaYNkCQQDrbxd+X9+/V6/w\nbqCVXt6DDsNUYs77yFLDwhISJtHfC2xGeP5hrigfeDPfYKeaRCpbs/j5+ZzgPVdn\nkaXdvrXXAkEA5SyvOxP+EmZxVBe3g+TsTBsmO5wnQLkYOyQqEMw5HLfXSprGwPEW\nIi8MCY2SDYb+UaLa43GKhoH3ekzwCfcs5QJAGgBs8dIY3gMLNVyic5zEqmjI/drj\nzT70lRYr9MFA0Idsb+QRBCy91avq3rLID+uTWgloaAM/ZiygKJoXXYQghQJBAJMB\nSdo0pdq5ueJ+YCqL0wOyuqCsNwWudZuiRBWIWu5QAxsJE4s6Wr9MvIT4OgLRYBuP\nwqb48yn6/nuGFMfftP0CQH8z7+3rOH+Imy+yedEJg0vXDMaTS5rACW/PPju8K2Ge\naXMydcPYbtqZueNPr6/fx+7xHBQMyqCX9xYdES6PFbw=\n-----END RSA PRIVATE KEY-----"
	rsaPrivateKey := "-----BEGIN PRIVATE KEY-----\nMIIEpAIBAAKCAQEA8UjDbjOPOsnTetzwrZn0EAfPu2rJqy+EPE4Vn1UZjLiHQOq9\nT4ATYCdQlYc4sKFqrgg5JopVqJDwm9D9QBOOhht7BwxgAiuM13CyUN3c3VXfPAL6\n7c6o5hgQRZKVME1bQ74SbrxSC599q0OVbTt/RCFn/lxnq8aVO6gFneOXoiIRAv4z\nDo6buobKvnlWn8AjqufDY/tmkbliiH/YygBSr5PlZFKv/xVwmemq4/yyoWr3mxJB\nP4pD/vYzDrbOwWDsuHne4t4lJqcEhfPf6d/X4vYirWGSgadMs/rFQeNgGxSTkSlg\nEIvgvEP/D5yNg9LYItXab7EXmeJndApyNjFotQIDAQABAoIBAQCuByJWCMAk0c9D\nzbmWLPV43nl5HxOn8dS5vkQgMLwF4E5Fh0nV3nWbI6AeyXIdC7+rGGlIH/yaYTf5\n2qnennxsw8NfmRBiYFDsTu9+HhegWMBzQ18DSMxPcUnDNG+AcDsESEJ5NkEi+v0k\nGtHbAE/RrxqXvsRThB688Xpjn/UNTYhGoJyWgCeQnX7Lb2f5ZuShxeTGITZZc7zX\nmVJlAxhB7ia8QjLvb3UP8Lj1LNRUbrf9ER1vWaxGkgJ2gVBtooU49SnaT1ytRIvF\nO7bebql0C1paJyIzeFVdOM+zkd3MtGSjLNeqx02u0RlDX/arWZzzfMIZUlXS/aI1\n+OEM+y5hAoGBAP4UWlf/PVipAkywGaEmVN7zhfDQwf22WVzFJ7t3ZUIGAuHLK9sg\nmwYsvL2wceeLdLdG5PDABn7Qg16iQomKc+IsrO9rpqFBRyhOxaqdh12x7uJEyE7K\nz4cFJWsuvYM2cT1r7yYelZcBO1IEYAfBd/fKS9SOuIHgVMOdaSyri4EpAoGBAPMb\npsdxcqqWZkqnscRdB7FwyL9Bu4yIXzQAXmFzp+CfvW5R7FyCkdpIUQ+45vp/8veO\nxoku0Yx6AMRmoML5s1H7xTVEfFxwOq+VxTujrDY4ISc9ExjAs8CmDa086VxfcMXL\ncmqPPiijq9fU510xbSRqqXrw7dgyXH7uiO1tViCtAoGBAPvxfqxngpNZ89fLf36W\nne3+mMB0oQhOO0qMjZuNBzFy+IShjLlbGZ4D2pxWTHyHgpMCj55TK52wcbPzPftY\n+RQ7rLg0jSS1sDO9+K/4SPSVrgVcoyvOW/7NEY+4518nKx722kt4ZGk6ZXIsuHS4\nWZeyV4NMipTut97JRjNvX4QRAoGAFX8VcT9L11+zLdjENEvRmVQSvC6wyc1pNazu\nq4lRh+A4HOArmUaGZRnN6Sj9a5sSvecnEQxNxTEojjVQ88ymmZI8rugeYwIdflcs\n+IgRez3dWJ3i6WZovpkwTmg7bkqWKzdBJ8EAyxEl6dDOoIP1VaoE0HUrYtmZKjQn\ncW8glYUCgYACTz6uuK44P29jWe573dMj/km+VS/yjz0QNmkLj9VdDj3Chl4njWhv\n+BDSZRslCBMGwtpHAEXns+v7pLhU6SyQItcsV4pEA/iNkM8g2ngUCMKyaP3H3BxS\nzaOXBae78oku9TsbCvgCT3G62NHpUw5von91B8IiPN0vm6KcMvZaRA==\n-----END PRIVATE KEY-----"

	block, _ := pem.Decode([]byte(rsaPrivateKey))
	rsa, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Fatal(err)
	}

	options := &dkim.SignOptions{
		//Domain:   "rtcrm.ru",
		Domain:   "ratuscrm.com",
		Selector: "dk1",
		Signer:   rsa,
	}

	var b bytes.Buffer
	if err := dkim.Sign(&b, r, options); err != nil {
		log.Fatal(err)
	}

	return b
}

func SendTestMail() {

	addr := "nkokorev@rus-marketing.ru"
	//mx, err := net.LookupMX("rus-marketing.ru")
	mx, err := net.LookupMX(addr)
	if err != nil {
		log.Fatal(err)
	}
	mxHost := strings.ToLower(mx[0].Host)
	fmt.Println("MX: ", mxHost)


	return
	/* Подключаемся к SMTP получателя */
	//c, err := smtp.Dial(mxHost + ":25")
	conn, err := net.DialTimeout("tcp", mxHost + ":25", 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	c, err := smtp.NewClient(conn,mxHost)

	/*c, err := smtp.Dial(mxHost + ":25")
	if err != nil {
		log.Fatal(err)
	}*/

	defer c.Close()

	fmt.Println("Connected!")

	tlc := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         mxHost,
	}
	if err := c.StartTLS(tlc); err != nil {
		log.Println("Не удалось запустить tlc")
	}

	// Проверяем адрес
	if err := c.Verify(addr);err != nil {
		fmt.Println("Адресок то поддельный: ", err)
	}

	return
	// Проверяем сервер
	/*if err := c.Hello("localhost"); err != nil {
		log.Fatal("Hello: ", err)
	}*/

	// Set the sender and recipient first
	if err := c.Mail("nk@ratuscrm.com"); err != nil {
		log.Fatal(err)
	}
	if err := c.Rcpt("nkokorev@rus-marketing.ru"); err != nil {
		log.Fatal(err)
	}



	header := make(map[string]string)
	header["From"] = "nk@ratuscrm.com"
	header["To"] = "nkokorev@rus-marketing.ru"
	header["Subject"] = "Тестируем сервер v2"
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=utf-8"
	header["Content-Transfer-Encoding"] = "base64"

	//body := "Это тестовое сообщение. Просьба не отвечать."
	body := "<H1>Header h1</H1><p>Это тестовое сообщение, просьба не отвечать</p>"
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))


	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		log.Fatal(err)
	}
	_, err = fmt.Fprintf(wc, message)
	if err != nil {
		log.Fatal(err)
	}
	err = wc.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Send the QUIT command and close the connection.
	err = c.Quit()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Message successful")
}

func SendEmailNew() (code int, message string, err error) {
	mx, err := net.LookupMX("rus-marketing.ru")
	if err != nil {
		log.Fatal(err)
	}
	mxRecord := mx[0].Host
	fmt.Println("Mx: ", mxRecord)

	//c, err := smtp.Dial("localhost:25")
	c, err := smtp.Dial(mxRecord + ":25")
	if err != nil {
		return
	}
	defer c.Quit() // make sure to quit the Client

	if err = c.Mail("nk@ratuscrm.com"); err != nil {
		return
	}

	if err = c.Rcpt("nkokorev@rus-marketing.ru"); err != nil {
		return
	}

	wc, err := c.Data()
	if err != nil {
		return
	}
	defer wc.Close() // make sure WriterCloser gets closed

	_, err = wc.Write([]byte("Hello my friend!"))
	if err != nil {
		return
	}

	code, message, err = c.Text.ReadResponse(0)
	return
}

func SendTestMail2() {
	from := mail.Address{"", "nk@ratuscrm.com"}
	to   := mail.Address{"", "nkokorev@rus-marketing.ru"}
	subj := "This is the email subject"
	body := "This is an example body.\n With two lines."

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subj

	// Setup message
	message := ""
	for k,v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Connect to the SMTP Server
	//servername := "smtp.example.tld:465"
	servername := "ratuscrm.ru:465"

	//host, _, _ := net.SplitHostPort(servername)
	host := "31.173.81.12"

	auth := smtp.PlainAuth("","username@example.tld", "password", host)

	// TLS config
	tlsconfig := &tls.Config {
		InsecureSkipVerify: true,
		ServerName: host,
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		log.Panic(err)
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		log.Panic(err)
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		log.Panic(err)
	}

	// To && From
	if err = c.Mail(from.Address); err != nil {
		log.Panic(err)
	}

	if err = c.Rcpt(to.Address); err != nil {
		log.Panic(err)
	}

	// Data
	w, err := c.Data()
	if err != nil {
		log.Panic(err)
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		log.Panic(err)
	}

	err = w.Close()
	if err != nil {
		log.Panic(err)
	}

	c.Quit()
}

func TestMailSMTP() {

	c, err := smtp.Dial("")
	if err != nil {
		log.Fatal(err)
	}
	/*c, err := smtp.NewClient(conn, "localhost")
	if err != nil {
		log.Fatal(err)
	}*/
	defer c.Close()
}

func Dial(addr string) (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", addr, nil) //smtp包中net.Dial()当使用ssl连接时会卡住
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}