package errors

import (
	"errors"
	_ "github.com/nkokorev/auth-server/locales"
	t "github.com/nkokorev/crm-go/locales"
)

var (
	EmailDoesNotExist	= errors.New(t.Trans(t.EmailDoesNotExist))
	EmailInvalidFormat	= errors.New(t.Trans(t.EmailInvalidFormat))

	UserUsernameIsTooShort = errors.New(t.Trans(t.UserUsernameIsTooShort))
	UserUsernameIsTooLong = errors.New(t.Trans(t.UserUsernameIsTooLong))
	UserUsernameForbiddenCharacters = errors.New(t.Trans(t.UserUsernameForbiddenCharacters))
)
