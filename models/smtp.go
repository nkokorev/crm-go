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

	byteMsg, err := signDKIM(message)
	if err != nil {
		log.Fatal(err)
		return err
	}

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

	_, err = fmt.Fprint(msg, byteMsg.String())
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

func signDKIM (mailString string) (*bytes.Buffer,error) {
	r := strings.NewReader(mailString)

	//rsaPrivateKey := "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDSw3hDW4hWLBZ2tLEZhl6lkXuBkxngTKoJm7Rim5pOGPuoxMekfKhj1egQA8Kh/+FnKVXJP/fsQpmoCGjxCdjC8dhUzUbZIj8OhBnMsa3uaAyOHNXnBWnZVfXSjtOQVfpJltt+SHy/CptXuX7TyvXZt65OdmKjvfHyvsByJEqdUwIDAQAB"
	//rsaPrivateKey := "-----BEGIN RSA PRIVATE KEY-----\nMIICXAIBAAKBgQDSw3hDW4hWLBZ2tLEZhl6lkXuBkxngTKoJm7Rim5pOGPuoxMek\nfKhj1egQA8Kh/+FnKVXJP/fsQpmoCGjxCdjC8dhUzUbZIj8OhBnMsa3uaAyOHNXn\nBWnZVfXSjtOQVfpJltt+SHy/CptXuX7TyvXZt65OdmKjvfHyvsByJEqdUwIDAQAB\nAoGAIDA8OMVM8CQxlhWIiqZr5Atw+lwV8pyix27hQMIU8eJ85MyQ1P041m5/z5pT\nalxi91dnw6GiYpHVV8VZCZ8AXJaP4f8DFOC5oAI6qe7xHtilnI3p8cItRVTEi6uU\nEamyh5fwZcVzq8yFAvIWc4P1bC5xJ8ce3P2QPVhpojIaYNkCQQDrbxd+X9+/V6/w\nbqCVXt6DDsNUYs77yFLDwhISJtHfC2xGeP5hrigfeDPfYKeaRCpbs/j5+ZzgPVdn\nkaXdvrXXAkEA5SyvOxP+EmZxVBe3g+TsTBsmO5wnQLkYOyQqEMw5HLfXSprGwPEW\nIi8MCY2SDYb+UaLa43GKhoH3ekzwCfcs5QJAGgBs8dIY3gMLNVyic5zEqmjI/drj\nzT70lRYr9MFA0Idsb+QRBCy91avq3rLID+uTWgloaAM/ZiygKJoXXYQghQJBAJMB\nSdo0pdq5ueJ+YCqL0wOyuqCsNwWudZuiRBWIWu5QAxsJE4s6Wr9MvIT4OgLRYBuP\nwqb48yn6/nuGFMfftP0CQH8z7+3rOH+Imy+yedEJg0vXDMaTS5rACW/PPju8K2Ge\naXMydcPYbtqZueNPr6/fx+7xHBQMyqCX9xYdES6PFbw=\n-----END RSA PRIVATE KEY-----"

	//rsaPrivateKey := "-----BEGIN RSA PRIVATE KEY-----\nMIICXAIBAAKBgQDSw3hDW4hWLBZ2tLEZhl6lkXuBkxngTKoJm7Rim5pOGPuoxMek\nfKhj1egQA8Kh/+FnKVXJP/fsQpmoCGjxCdjC8dhUzUbZIj8OhBnMsa3uaAyOHNXn\nBWnZVfXSjtOQVfpJltt+SHy/CptXuX7TyvXZt65OdmKjvfHyvsByJEqdUwIDAQAB\nAoGAIDA8OMVM8CQxlhWIiqZr5Atw+lwV8pyix27hQMIU8eJ85MyQ1P041m5/z5pT\nalxi91dnw6GiYpHVV8VZCZ8AXJaP4f8DFOC5oAI6qe7xHtilnI3p8cItRVTEi6uU\nEamyh5fwZcVzq8yFAvIWc4P1bC5xJ8ce3P2QPVhpojIaYNkCQQDrbxd+X9+/V6/w\nbqCVXt6DDsNUYs77yFLDwhISJtHfC2xGeP5hrigfeDPfYKeaRCpbs/j5+ZzgPVdn\nkaXdvrXXAkEA5SyvOxP+EmZxVBe3g+TsTBsmO5wnQLkYOyQqEMw5HLfXSprGwPEW\nIi8MCY2SDYb+UaLa43GKhoH3ekzwCfcs5QJAGgBs8dIY3gMLNVyic5zEqmjI/drj\nzT70lRYr9MFA0Idsb+QRBCy91avq3rLID+uTWgloaAM/ZiygKJoXXYQghQJBAJMB\nSdo0pdq5ueJ+YCqL0wOyuqCsNwWudZuiRBWIWu5QAxsJE4s6Wr9MvIT4OgLRYBuP\nwqb48yn6/nuGFMfftP0CQH8z7+3rOH+Imy+yedEJg0vXDMaTS5rACW/PPju8K2Ge\naXMydcPYbtqZueNPr6/fx+7xHBQMyqCX9xYdES6PFbw=\n-----END RSA PRIVATE KEY-----\n"
	//rsaPrivateKey := "-----BEGIN PRIVATE KEY-----\nMIIBCAKBgQDQZquP/DQkKtx2J2Q0XQSgusZJkfClLPRZfxKVHhOOGIkK4aeitocO\nRsg8L7cgHWSSvIwI42vjKrUVmF8xYdf2KgkEaRHcDkWalH7ufVYzGzF/0sioRLVp\n+c5zJhPes0TWWfqe9VJ7nc30qdESeQAnkQJq9t0FlEVKOKt7ligEIwKBgQDIEuqA\ndwnLpMVKGnT62NaPMReNd2YtjUJjBG9q1PfmnUaNLxQYtwWhffInnlHWlR3b8VBk\nZfuiUzG+u282lBEbYPaiifD5lO8m0uG0K2Zofcf820j/xKQo/yeZ7T02RasOJpeH\nYwb960JuwkBOF6yqY6GTUL2FRLj0ACmW5eL3Lw==\n-----END PRIVATE KEY-----"
	rsaPrivateKey := "-----BEGIN RSA PRIVATE KEY-----\nMIICXQIBAAKBgQCwy7WZIg2haroLTj14GS7MVeLyR0RE7hkhdPYVjdKlUlaJeun5\nlwp7//QcQmZPu9O7e46mTD+CE6srCVyKWCSeUlAVwcV7GT7A9VKnPPiGgAs26Hqz\nAuGwhER3l+lT1arVTbRu7E6shBoWROwPAqZPPp+jctL79CEta5U2ICduHQIDAQAB\nAoGAE+aKRXd400+hK36eGrOy+ds9FYqCG8Q1Xfe9b4WsTWGsTgNg7PBchMK15qxu\nudDpr3PkBcIVb/3oyYpfOU9cp6mgXk557OxqfPNyNwRO/o/6/IiEpFFrk8jJxoc3\nmoa9Lh1hM/lsSGryp83L1vBUTs3tXIGo+uBBHnLaH33dFF0CQQDhizg/xVAhR4he\n8Q/uSP5Cgf/Viwevluxpz2R4WrGro5XRyLvEoXb+gPG9NqjT62N7jHX1lBxFpFPT\n/zh1BADLAkEAyKtTmww6/ULKTijfBOhp+w/O4TOWbq0JSZBXAGPI6jh+73gGNf/x\n+55kMYUjIaxpIkILsDTlQrO5kBIBarX3twJAHtXp2s4fJm2hN1m909Ym7PDZCVj4\ntAjuSYkRM2My50R2Nzg6c6efnSwD4NqYOmD0OO/7MJgPRXYx/8nk7hqeAQJBAJ96\n8h42cSdYjpnhh6VJ5PigTqXSLwtUwB3T9iEcLNBhCBjfhegiurlj33MvwYUAlimg\n3dMzpsUFO0PR24hoiC8CQQDP1kDw2zzA8dwGFjbBPqFfN5uVcbwzq1tRjdM1mkp8\nwJB/anwuIRNIE/PDCvi4MEmW7p7FkfbHOZOSgYXbIK3k\n-----END RSA PRIVATE KEY-----"

	block, _ := pem.Decode([]byte(rsaPrivateKey))
	rsa, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	if err != nil {
		return nil, err
	}

	options := &dkim.SignOptions{
		//Domain:   "rtcrm.ru",
		Domain:   "mta1.ratuscrm.com",
		Selector: "dk1",
		Signer:   rsa,
	}

	var b bytes.Buffer
	if err := dkim.Sign(&b, r, options); err != nil {
		log.Fatal(err)
	}

	return &b, nil
}
