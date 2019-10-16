package utils

import (
	"errors"
	"fmt"
	_ "github.com/nkokorev/auth-server/locales"
	t "github.com/nkokorev/crm-go/locales"
	"net"
	"net/smtp"
	"regexp"
	"strings"
	"time"
)

type SmtpError struct {
	Err error
}

func (e SmtpError) Error() string {
	return e.Err.Error()
}

func (e SmtpError) Code() string {
	return e.Err.Error()[0:3]
}

func NewSmtpError(err error) SmtpError {
	return SmtpError{
		Err: err,
	}
}

const forceDisconnectAfter = time.Second * 5

var (
	ErrBadFormat        = errors.New("invalid format")
	ErrUnresolvableHost = errors.New("unresolvable host")

	emailRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	userRegexp = regexp.MustCompile("^[a-zA-Z0-9!#$%&'*+/=?^_`{|}~.-]+$")
	hostRegexp = regexp.MustCompile("^[^\\s]+\\.[^\\s]+$")
	// As per RFC 5332 secion 3.2.3: https://tools.ietf.org/html/rfc5322#section-3.2.3
	// Dots are not allowed in the beginning, end or in occurances of more than 1 in the email address
	userDotRegexp = regexp.MustCompile("(^[.]{1})|([.]{1}$)|([.]{2,})")
)

func VerifyEmail(email string, opt_deep... bool) (error Error) {
	deep := false
	if len(opt_deep) > 0 {
		deep = opt_deep[0]
	}
	error = ValidateFormat(email)


	if !error.HasErrors() {
		if deep {
			error = ValidateEmailDeepHost(email)
		} else {
			error = ValidateHost(email)
		}
	}
	return
}

func ValidateFormat(email string) (error Error) {

	if len(email) < 6 || len(email) > 254 {
		error.AddErrors("email", t.Trans(t.EmailInvalidFormat) )
		return
	}

	at := strings.LastIndex(email, "@")
	if at <= 0 || at > len(email)-3 {
		error.AddErrors("email", t.Trans(t.EmailInvalidFormat) )
		return
	}

	user := email[:at]
	host := email[at+1:]

	if len(user) > 64 {
		error.AddErrors("email", t.Trans(t.EmailInvalidFormat) )
		return
	}

	if userDotRegexp.MatchString(user) || !userRegexp.MatchString(user) || !hostRegexp.MatchString(host) {
		error.AddErrors("email", t.Trans(t.EmailInvalidFormat) )
		return
	}

	switch host {
	case "localhost", "example.com":
		error.AddErrors("email", t.Trans(t.EmailInvalidFormat) )
		return
		//return nil // хоть это и валидный адрес, лесом....
	}

	/*if !emailRegexp.MatchString(email) {
		error.AddErrors("email", t.Trans(t.EmailInvalidFormat) )
	}*/
	return
}

func ValidateHost(email string) (error Error) {

	at := strings.LastIndex(email, "@")
	host := email[at+1:]

	if _, err := net.LookupMX(host); err != nil {
		if _, err := net.LookupIP(host); err != nil {
			// Only fail if both MX and A records are missing - any of the
			// two is enough for an email to be deliverable
			error.AddErrors("email", t.Trans(t.EmailUnresolvableHost) )
			return
		}
	}
	return
}

func ValidateEmailDeepHost(email string) (error Error) {
	_, host := split(email)
	mx, err := net.LookupMX(host)
	if err != nil {
		error.AddErrors("email", t.Trans(t.EmailDoesNotExist) )
		return
		//return ErrUnresolvableHost
	}

	client, err := DialTimeout(fmt.Sprintf("%s:%d", mx[0].Host, 25), forceDisconnectAfter)
	if err != nil {
		error.AddErrors("email", t.Trans(t.EmailDoesNotExist) )
		return
		//return NewSmtpError(err)
	}
	defer client.Close()

	err = client.Hello("checkmail.me")
	if err != nil {
		error.AddErrors("email", t.Trans(t.EmailDoesNotExist) )
		return
		//return NewSmtpError(err)
	}
	err = client.Mail("lansome-cowboy@gmail.com")
	if err != nil {
		error.AddErrors("email", t.Trans(t.EmailDoesNotExist) )
		return
		//return NewSmtpError(err)
	}
	err = client.Rcpt(email)
	if err != nil {
		error.AddErrors("email", t.Trans(t.EmailDoesNotExist) )
		return
		//return NewSmtpError(err)
	}
	return
}

// DialTimeout returns a new Client connected to an SMTP server at addr.
// The addr must include a port, as in "mail.example.com:smtp".
func DialTimeout(addr string, timeout time.Duration) (*smtp.Client, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}

	t := time.AfterFunc(timeout, func() { conn.Close() })
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
