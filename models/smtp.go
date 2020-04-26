package models

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
)

//ex: SendMail("127.0.0.1:25", (&mail.Address{"from name", "from@example.com"}).String(), "Email Subject", "message body", []string{(&mail.Address{"to name", "to@example.com"}).String()})
func SendMail(addr, from, subject, body string, to []string) error {
	r := strings.NewReplacer("\r\n", "", "\r", "", "\n", "", "%0a", "", "%0d", "")

	c, err := smtp.Dial(addr)
	if err != nil {
		fmt.Println("Error in dial()")
		return err
	}
	defer c.Close()
	if err = c.Mail(r.Replace(from)); err != nil {
		return err
	}
	for i := range to {
		to[i] = r.Replace(to[i])
		if err = c.Rcpt(to[i]); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	msg := "To: " + strings.Join(to, ",") + "\r\n" +
		"From: " + from + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"Content-Transfer-Encoding: base64\r\n" +
		"\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

func SendTestMail() {

	//ns, err := net.LookupNS("rus-marketing.ru")
	mx, err := net.LookupMX("rus-marketing.ru")
	if err != nil {
		log.Fatal(err)
	}
	mxHost := mx[0].Host
	fmt.Println("NS: ", mxHost)

	//c, err := smtp.Dial("dns2.yandex.net:465")
	//c, err := net.DialTimeout("tcp","dns2.yandex.net.:25", 10*time.Second)
	//c, err := smtp.Dial("tcp","mx.yandex.net.:25", 10*time.Second)
	//conn, err := net.DialTimeout("tcp", mxRecord + ":25", 5*time.Second)
	/*conn, err := net.DialTimeout("tcp", mxRecord + ":25", 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	*/
	c, err := smtp.Dial(mxHost + ":25")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	fmt.Println("Connected!")

	tlc := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         mxHost,
	}
	if err := c.StartTLS(tlc); err != nil {
		log.Println("Не удалось запустить tlc")
	}

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
	header["Subject"] = "Test message!"
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	body := ""
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