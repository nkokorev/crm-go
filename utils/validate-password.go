package utils

import (
	"errors"
	"unicode"
)

func VerifyPassword(pwd string) error {

	if len([]rune(pwd)) == 0 {
		return errors.New("Поле необходимо заполнить")
	}
	if len([]rune(pwd)) < 6 {
		return errors.New("Слишком короткий пароль")
	}

	if len([]rune(pwd)) > 25 {
		return errors.New("Слишком длинный пароль")
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
		return errors.New("Слишком простой пароль: проверьте наличие обязательных символов.")
	}

	return nil
}
