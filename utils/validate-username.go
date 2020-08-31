package utils

import (
	"errors"
	"regexp"
)

// проверяет имя пользователя на соответствие правилам. Не проверяет уникальность
func VerifyUsername(username *string) error {

	if username == nil { return nil }

	if len(*username) == 0 {
		return errors.New("Поле необходимо заполнить")
	}

	if len(*username) < 3 {
		return errors.New("Имя пользователя слишком короткое")
	}

	if len(*username) > 25 || len([]rune(*username)) > 25 {
		return errors.New("Имя пользователя слишком длинное")
	}

	var rxUsername = regexp.MustCompile("^[a-zA-Z0-9,\\-_]+$")

	if !rxUsername.MatchString(*username) {
		return errors.New("Используйте только a-z,A-Z,0-9 а также символ -")
	}

	return nil
}
