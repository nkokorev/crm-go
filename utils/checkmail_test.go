package utils

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

