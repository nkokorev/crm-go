package errors

import (
	"errors"
	_ "github.com/nkokorev/auth-server/locales"
	t "github.com/nkokorev/crm-go/locales"
)

var (
	AccountFailedToCreate = errors.New(t.Trans(t.AccountFailedToCreate))
	AccountDeletionError = errors.New(t.Trans(t.AccountDeletionError))

	EmailDoesNotExist	= errors.New(t.Trans(t.EmailDoesNotExist))
	EmailInvalidFormat	= errors.New(t.Trans(t.EmailInvalidFormat))

	UserFailedToCreate	= errors.New(t.Trans(t.UserFailedToCreate))
	UserUsernameIsTooShort = errors.New(t.Trans(t.UserUsernameIsTooShort))
	UserUsernameIsTooLong = errors.New(t.Trans(t.UserUsernameIsTooLong))
	UserUsernameIsRequired = errors.New(t.Trans(t.UserUsernameIsRequired))
	UserUsernameForbiddenCharacters = errors.New(t.Trans(t.UserUsernameForbiddenCharacters))
	UserDeletionErrorNotID = errors.New(t.Trans(t.UserDeletionErrorNotID))
	UserDeletionErrorHasAccount = errors.New(t.Trans(t.UserDeletionErrorHasAccount))

	UserPasswordIsRequired = errors.New(t.Trans(t.UserPasswordIsRequired))
	UserPasswordIsTooShort = errors.New(t.Trans(t.UserPasswordIsTooShort))
	UserPasswordIsTooLong = errors.New(t.Trans(t.UserPasswordIsTooLong))
	UserPasswordIsTooSimple = errors.New(t.Trans(t.UserPasswordIsTooSimple))

	UserRoleAppendFailed = errors.New(t.Trans(t.UserRoleAppendFailed))
	UserRoleRemoveFailed = errors.New(t.Trans(t.UserRoleRemoveFailed))


	RoleDeletionError = errors.New(t.Trans(t.RoleDeletionError))
	RoleDeletedFailedHasUsers = errors.New(t.Trans(t.RoleDeletedFailedHasUsers))
	RoleChangeOwnerRoleFailed = errors.New(t.Trans(t.RoleChangeOwnerRoleFailed))

	InputIsTooLong = errors.New(t.Trans(t.InputIsTooLong))
	InputIsTooShort = errors.New(t.Trans(t.InputIsTooShort))
	InputIsRequired = errors.New(t.Trans(t.InputIsRequired))


)
