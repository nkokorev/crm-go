package utils

import (
	"testing"
)

type emails struct {
	email string
	result bool
}

var tests = []emails{
	{ "kokorevn@gmail.com", true },
	{ "nkokorev@rus-marketing.ru", true },
	{ "asedsa23443_sada.", false },
	{ "no-user@rus-marketing.ru", true },
}

func TestVerifyEmail(t *testing.T) {
	for _, pair := range tests {
		err := VerifyEmail(pair.email, false)

		if err != nil && pair.result == true {
			t.Error(
				"For", pair.email,
				"expected", pair.result,
				"got", ( !pair.result),
				"Error: ", err.Error(),
			)
		}
	}
}
