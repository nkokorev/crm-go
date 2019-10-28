package utils

import (
	_ "github.com/nkokorev/auth-server/locales"
	e "github.com/nkokorev/crm-go/errors"
	"regexp"
	"unicode"
)

func VerifyPassword(pwd string) error {

	if len([]rune(pwd)) == 0 {
		return e.UserPasswordIsRequired
	}
	if len([]rune(pwd)) < 6 {
		return e.UserPasswordIsTooShort
	}

	if len([]rune(pwd)) > 25 {
		return e.UserPasswordIsTooLong
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
		return e.UserPasswordIsTooSimple
	}

	return nil
}

// проверяет имя пользователя на соответствие правилам. Не проверяет уникальность
func VerifyUsername(username string) error {

	if len(username) == 0 {
		return e.UserUsernameIsRequired
	}
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
