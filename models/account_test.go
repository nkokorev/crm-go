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
	err := test_user_1.Create()
	if err != nil {
		t.Error(err.Error())
	} else {
		defer test_user_1.Delete()
	}

	test_account_1 := Account {	Name:"Account_Test",}
	err = test_account_1.Create(&test_user_1)
	if err != nil {
		t.Error(err.Error())
	} else {
		defer test_account_1.Delete()
	}

	temp_account := Account{}
	if base.GetDB().First(&temp_account, test_account_1.ID).RecordNotFound() {
		t.Errorf("Cant find created account: %v", test_account_1.Name)
	}

	// проверим, что в аккаунт был добавлен владелец аккаунта
	if base.GetDB().Model(&test_user_1).Where("account_id = ?", test_account_1.ID).Association("Accounts").Count() != 1 {
		t.Error("Владелец аккаунта не добавлен в список пользователей аккаунта")
	}

	// проверим, что нам не дадут удалить пользователя, если у него есть аккаунты
	err = test_user_1.Delete()
	if err == nil {
		t.Error("Удалось удалить пользователя, при существующем аккаунте")
	}
	if base.GetDB().First(&temp_account, test_account_1.ID).RecordNotFound() {
		t.Errorf("аккаунт был удален, хотя есть ограничение внешнего ключа: %v", test_account_1.Name)
	}

	// удаляем аккаунт
	err = test_account_1.Delete()
	if err != nil {
		t.Error(err.Error())
	}

	// удаляем пользователя (должен удалиться)
	err = test_user_1.Delete()
	if err != nil {
		t.Error("Неудалось удалить пользователя, хотя аккаунт был удален.")
	}

}

