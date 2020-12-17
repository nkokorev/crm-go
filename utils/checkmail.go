package utils

import (
	"errors"
	"fmt"
	"net"
	"net/smtp"
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

func EmailValidation(email string) error {

	// todo edit!
	if len(email) < 6 {
		return errors.New("Email-адрес слишком короткий")
	}
	// ограничиваем 60 вместо 254
	if len(email) > 60 {
		return errors.New("Email-адрес слишком длинный")
	}

	at := strings.LastIndex(email, "@")
	if at <= 0 || at > len(email)-3 {
		return errors.New("Не верный формат email-адреса")
	}

	user := email[:at]
	host := email[at+1:]

	if len(user) > 64 {
		return errors.New("Email-адрес указан не верно")
	}

	if userDotRegexp.MatchString(user) || !userRegexp.MatchString(user) || !hostRegexp.MatchString(host) {
		return errors.New("Не верный формат email-адреса")
	}

	switch host {
	case "localhost", "example.com":
		return errors.New("Данный email-адрес не существует")
		//return nil // хоть это и валидный адрес, лесом....
	}

	return nil
}

func EmailDeepValidation(email string) error {

	if err := EmailValidation(email); err !=nil {
		return err
	}
	_, host := split(email)

	mx, err := net.LookupMX(host)
	if err != nil {
		return errors.New(fmt.Sprintf("Несуществующий домен почты: %v", host))
	}

	client, err := DialTimeout(fmt.Sprintf("%s:%d", mx[0].Host, 25), forceDisconnectAfter)
	if err != nil {
		return errors.New(fmt.Sprintf("Почтовый сервер не отвечает: %v", mx[0].Host))
	}
	defer func() {
		if err := client.Close();err!=nil {
			// тут можно какой-то лог записать
			return
		}
	}()

	err = client.Hello("checkmail.me")
	if err != nil {
		return errors.New( "Похоже, почтовый сервер не готов принять почту")
		//return NewSmtpError(err)
	}

	err = client.Mail("lansome-cowboy@gmail.com")
	if err != nil {
		return errors.New("Почтовый адрес не может принять почту")
	}

	err = client.Rcpt(email)
	if err != nil {
		return errors.New("Похоже, почтовый адрес не сущесвует")
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

func ValidateHostEmail(email string) error {

	at := strings.LastIndex(email, "@")
	host := email[at+1:]

	if _, err := net.LookupMX(host); err != nil {
		if _, err := net.LookupIP(host); err != nil {
			// Only fail if both MX and A records are missing - any of the
			// two is enough for an email to be deliverable
			//error.AddErrors("email", t.Trans(t.EmailUnresolvableHost) )
			return errors.New("Неверный формат")
		}
	}
	return nil
}