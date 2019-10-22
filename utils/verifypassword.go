package utils

import (
	"errors"
	_ "github.com/nkokorev/auth-server/locales"
	e "github.com/nkokorev/crm-go/errors"
	t "github.com/nkokorev/crm-go/locales"
	"regexp"
	"unicode"
)

// todo может сделать больше переменных и затрансовать их сразу?
var (
	UserPasswordIsTooShort	= errors.New(t.Trans(t.UserPasswordIsTooShort))
	UserPasswordRequired	= errors.New(t.Trans(t.UserPasswordRequired))
	UserPasswordIsTooLong	= errors.New(t.Trans(t.UserPasswordIsTooLong))
	UserPasswordIsTooSimple	= errors.New(t.Trans(t.UserPasswordIsTooSimple))
)

func VerifyPassword(pwd string) error {

	if len(pwd) < 6 {
		return UserPasswordIsTooShort
	}

	if len(pwd) == 0 {
		return UserPasswordRequired
	}

	if len(pwd) > 25 {
		return UserPasswordIsTooLong
	}

	letters := 0
	var number, upper, special bool
	for _, c := range pwd {
		switch {
		case unicode.IsNumber(c):
			number = true
		case unicode.IsUpper(c):
			upper = true
			letters++
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			special = true
		case unicode.IsLetter(c) || c == ' ':
			letters++
		default:
			//return false, false, false, false
		}
	}

	if ! (number && upper && special && letters >= 5) {
		return UserPasswordIsTooSimple
	}

	return nil
}

func VerifyUsername(username string) error {

	if len(username) < 3 {
		return e.UserUsernameIsTooShort
	}
	if len(username) > 25 || len([]rune(username)) > 25 {
		return e.UserUsernameIsTooLong
	}

	var rxUsername = regexp.MustCompile("^[a-zA-Z0-9,-,_]+$")

	if !rxUsername.MatchString(username) {
		return e.UserUsernameForbiddenCharacters
	}


	return nil
}
