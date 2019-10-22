package models

import (
	"github.com/nkokorev/crm-go/database/base"
	"testing"
)

func TestExistAccountTable(t *testing.T) {
	db := base.GetDB()
	if !db.HasTable(&Account{}) {
		tableName := db.NewScope(&Account{}).GetModelStruct().TableName(db)
		t.Errorf("%v table is not exist!",tableName)
	}
}

func TestAccount_Create(t *testing.T) {

	test_user_1 := User{
		Username:"user_test",
		Email: "testmail@ratus-dev.ru",
		Name:"РеальноеИмя",
		Surname:"РеальнаяФамилия",
		Patronymic:"РеальноеОтчество",
		Password: "qwerty123#Aa",
	}

	myErr := test_user_1.Create()
	if myErr.HasErrors() {
		t.Error(myErr.Message)
		for k, r := range myErr.GetErrors() {
			t.Errorf("%s | %s", k,r)
		}
		return
	} else {
		defer test_user_1.Delete()
	}

	test_account_1 := Account {
		Name:"Account_Test",
	}

	myErr = test_account_1.Create(&test_user_1)
	if myErr.HasErrors() {
		t.Error(myErr.Message)
		for k, r := range myErr.GetErrors() {
			t.Errorf("%s | %s", k,r)
		}
		return
	} else {
		defer test_account_1.Delete()
	}

	temp_account := Account{}
	if base.GetDB().First(&temp_account, test_account_1.ID).RecordNotFound() {
		t.Errorf("Cant find created account: %v", test_account_1.Name)
	}

	// проверим, что при удалении пользователя, аккаунт тоже удалится
	// проверим, что нам не дадут удалить пользователя, если у него есть аккаунты
	myErr = test_user_1.Delete()
	if !myErr.HasErrors() {
		t.Error("Удалось удалить пользователя, при существующем аккаунте")
		return
	}
	if base.GetDB().First(&temp_account, test_account_1.ID).RecordNotFound() {
		t.Errorf("аккаунт был удален, хотя есть ограничение внешнего ключа: %v", test_account_1.Name)
	}

	// удаляем аккаунт
	myErr = test_account_1.Delete()
	if myErr.HasErrors() {
		t.Error(myErr.Message)
		for k, r := range myErr.GetErrors() {
			t.Errorf("%s | %s", k,r)
		}
		return
	}

	// удаляем пользователя (должен удалиться)
	myErr = test_user_1.Delete()
	if myErr.HasErrors() {
		t.Error("Неудалось удалить пользователя, хотя аккаунт был удален.")
		return
	}


}

