package utils

import (
	"fmt"
	_ "github.com/nkokorev/auth-server/locales"
	e "github.com/nkokorev/crm-go/errors"
	"net"
	"net/smtp"
	"os"
	"regexp"
	"strings"
	"time"
)

const forceDisconnectAfter = time.Second * 5

var (
	userRegexp = regexp.MustCompile("^[a-zA-Z0-9!#$%&'*+/=?^_`{|}~.-]+$")
	hostRegexp = regexp.MustCompile("^[^\\s]+\\.[^\\s]+$")
	// As per RFC 5332 secion 3.2.3: https://tools.ietf.org/html/rfc5322#section-3.2.3
	// Dots are not allowed in the beginning, end or in occurances of more than 1 in the email address
	userDotRegexp = regexp.MustCompile("(^[.]{1})|([.]{1}$)|([.]{2,})")
)

func VerifyEmail(email string, opt_deep... bool) error {

	deep := false
	if len(opt_deep) > 0 {
		deep = opt_deep[0]
	}

	err := ValidateFormat(email)
	if err != nil {
		return  err
	}

	if (os.Getenv("http_dev") == "true") {
		return nil
	}

	if deep {
		err = ValidateEmailDeepHost(email)
	} else {
		err = ValidateHost(email)
	}

	return err
}

func ValidateFormat(email string) error {

	if len(email) < 6 || len(email) > 254 {
		return e.EmailInvalidFormat
	}

	at := strings.LastIndex(email, "@")
	if at <= 0 || at > len(email)-3 {
		return e.EmailInvalidFormat
	}

	user := email[:at]
	host := email[at+1:]

	if len(user) > 64 {
		return e.EmailInvalidFormat
	}

	if userDotRegexp.MatchString(user) || !userRegexp.MatchString(user) || !hostRegexp.MatchString(host) {
		return e.EmailInvalidFormat
	}

	switch host {
	case "localhost", "example.com":
		return e.EmailInvalidFormat
		//return nil // хоть это и валидный адрес, лесом....
	}

	/*if !emailRegexp.MatchString(email) {
		error.AddErrors("email", t.Trans(t.EmailInvalidFormat) )
	}*/
	return nil
}

func ValidateHost(email string) error {

	at := strings.LastIndex(email, "@")
	host := email[at+1:]

	if _, err := net.LookupMX(host); err != nil {
		if _, err := net.LookupIP(host); err != nil {
			// Only fail if both MX and A records are missing - any of the
			// two is enough for an email to be deliverable
			//error.AddErrors("email", t.Trans(t.EmailUnresolvableHost) )
			return e.EmailInvalidFormat
		}
	}
	return nil
}

func ValidateEmailDeepHost(email string) error {
	_, host := split(email)

	mx, err := net.LookupMX(host)
	if err != nil {
		return e.EmailDoesNotExist
	}

	client, err := DialTimeout(fmt.Sprintf("%s:%d", mx[0].Host, 25), forceDisconnectAfter)
	if err != nil {
		return e.EmailDoesNotExist
	}
	defer func() {
		if err := client.Close();err!=nil {
			// тут можно какой-то лог записать
			return
		}
	}()

	err = client.Hello("checkmail.me")
	if err != nil {
		return e.EmailDoesNotExist
		//return NewSmtpError(err)
	}

	err = client.Mail("lansome-cowboy@gmail.com")
	if err != nil {
		return e.EmailDoesNotExist
	}

	err = client.Rcpt(email)
	if err != nil {
		return e.EmailDoesNotExist
	}

	return nil
}

// DialTimeout returns a new Client connected to an SMTP server at addr.
// The addr must include a port, as in "mail.example.com:smtp".
func DialTimeout(addr string, timeout time.Duration) (*smtp.Client, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}

	t := time.AfterFunc(timeout, func() {
		if err := conn.Close(); err != nil {
			// todo there...
		}
	})
	defer t.Stop()

	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}

func split(email string) (account, host string) {
	i := strings.LastIndexByte(email, '@')
	account = email[:i]
	host = email[i+1:]
	return
}
