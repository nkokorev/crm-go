package utils

import (
	_ "github.com/nkokorev/auth-server/locales"
	t "github.com/nkokorev/crm-go/locales"
	"unicode"
)

func VerifyPassword(pwd string) (error Error) {

	if len(pwd) < 6 {
		error.AddErrors("password", t.Trans(t.UserPasswordIsTooShort) )
	}

	if len(pwd) == 0 {
		error.AddErrors("password", t.Trans(t.UserPasswordRequired) )
	}

	if len(pwd) > 25 {
		error.AddErrors("password", t.Trans(t.UserPasswordIsTooLong) )
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
		error.AddErrors("password", t.Trans(t.UserPasswordIsTooSimple) )
	}

	return
}
