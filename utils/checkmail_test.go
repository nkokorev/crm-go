package utils

import (
	"testing"
)

type emails struct {
	email string
	result bool
	http_dev bool // тестить ли при дев формате
}

var tests = []emails{
	{ "kokorevn@gmail.com", true, true },
	{ "nkokorev@rus-marketing.ru", true, true },
	{ "asedsa23443_sada.", false, true },
	{ "asedsa23443_sada.@fdsfdfds.com", false, true },
	{ "no-user@rus-marketing.ru", false, false }, // такого пользователя нет
}

func TestVerifyEmail(t *testing.T) {
	for _, pair := range tests {
		err := VerifyEmail(pair.email, true)

		// обходим тесты, которые требуют http соединения
		if ! pair.http_dev {
			break
		}

		if pair.result != (err == nil) {
			t.Error(
				"For", pair.email,
				"expected", pair.result,
				"got", ( !pair.result),
				"Error: ", err,
			)
		}
	}
}
