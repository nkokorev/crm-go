package smtp

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/smtp"
)

/* Рабочий вариант */

func WorkV1() {

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
