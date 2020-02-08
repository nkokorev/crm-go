package models

import (
	"github.com/nkokorev/crm-go/utils"
	"testing"
)

func TestAccount_ValidateInputs(t *testing.T) {

	account := Account{Name:""}
	if err := account.ValidateInputs(); err == nil {
		t.Fatal("Validate account without name")
	}

	account.Name = utils.RandStringBytes(42)
	if err := account.ValidateInputs(); err == nil {
		t.Fatal("Validate account with very long name")
	}

	account.Name = utils.RandStringBytes(10)
	if err := account.ValidateInputs(); err != nil {
		t.Fatal("No validate account with shot name")
	}

	account.Website = utils.RandStringBytes(256)
	if err := account.ValidateInputs(); err == nil {
		t.Fatal("Validate account with very long website name")
	}

	account.Website = utils.RandStringBytes(50)
	if err := account.ValidateInputs(); err != nil {
		t.Fatal("No Validate account with norm website name")
	}

	account.Type = utils.RandStringBytes(256)
	if err := account.ValidateInputs(); err == nil {
		t.Fatal("Validate account with very long type")
	}

	account.Type = utils.RandStringBytes(50)
	if err := account.ValidateInputs(); err != nil {
		t.Fatal("No Validate account with norm type")
	}

}

func TestAccount_createAccount(t *testing.T) {

	// 1. Аккаунт не должен создаваться без вводных данных
	testAccount, err := createAccount(Account{})
	if err == nil && testAccount != nil {
		defer testAccount.HardDelete()
		t.Fatal("Created account, but name is null")
	}

	outAccount, err := createAccount( Account{Name: "Test account"} )
	if err != nil || outAccount == nil {
		t.Fatal("Cant create account without name")
	}

	defer outAccount.HardDelete()

}
