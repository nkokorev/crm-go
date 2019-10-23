package errors

import (
	"errors"
	_ "github.com/nkokorev/auth-server/locales"
	t "github.com/nkokorev/crm-go/locales"
)

var (
	AccountDeletionError = errors.New(t.Trans(t.AccountDeletionError))

	EmailDoesNotExist	= errors.New(t.Trans(t.EmailDoesNotExist))
	EmailInvalidFormat	= errors.New(t.Trans(t.EmailInvalidFormat))

	UserUsernameIsTooShort = errors.New(t.Trans(t.UserUsernameIsTooShort))
	UserUsernameIsTooLong = errors.New(t.Trans(t.UserUsernameIsTooLong))
	UserUsernameIsRequired = errors.New(t.Trans(t.UserUsernameIsRequired))
	UserUsernameForbiddenCharacters = errors.New(t.Trans(t.UserUsernameForbiddenCharacters))

	UserPasswordIsRequired = errors.New(t.Trans(t.UserPasswordIsRequired))
	UserPasswordIsTooShort = errors.New(t.Trans(t.UserPasswordIsTooShort))
	UserPasswordIsTooLong = errors.New(t.Trans(t.UserPasswordIsTooLong))
	UserPasswordIsTooSimple = errors.New(t.Trans(t.UserPasswordIsTooSimple))


)
