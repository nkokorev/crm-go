package smtp

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
)


func RunSMTP() error {

	auth := sasl.NewPlainClient("", "mex388", "")

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	mailTo := "nkokorev@rus-marketing.ru"

	msg := strings.NewReader("To: " + mailTo + "\r\n" +
		"Subject: discount Gophers!\r\n" +
		"\r\n" +
		"This is the email body.\r\n")

	host, err := GetHostMX(mailTo)
	if err != nil {
		return err
	}

	err = smtp.SendMail(*host, auth, "info@ratuscrm.com", []string{mailTo}, msg)
	if err != nil {
		log.Fatal(err)
	}


}

func GetHostMX(address string) (*string, error){
	ports := []int{25, 2525, 587}

	host := strings.Split(address, "@")[1]
	mxs, err := net.LookupMX(host)
	if err != nil {
		return nil, err
	}

	for i := range mxs {
		for j := range ports {
			server := strings.TrimSuffix(mxs[i].Host, ".")
			hostPort := fmt.Sprintf("%s:%d", server, ports[j])

			_, err := net.DialTimeout("tcp", hostPort, 5*time.Second)
			if err != nil {
				if j == len(ports)-1 {
					return nil, err
				}

				continue
			}

			return &hostPort, nil
		}
	}

	return nil, errors.New("Не удалось найти нужный порт")
}