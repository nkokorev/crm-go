package models

import (
	"github.com/nkokorev/crm-go/database/base"
	"testing"
)

func TestExistAccountUserTable(t *testing.T) {
	db := base.GetDB()
	if !db.HasTable(&AccountUser{}) {
		tableName := db.NewScope(&AccountUser{}).GetModelStruct().TableName(db)
		t.Errorf("%v table is not exist!",tableName)
	}
}

func TestAccountUser_GetAccountUser(t *testing.T) {
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

	test_acc_user_1 := AccountUser{}
	err = test_acc_user_1.GetAccountUser(test_user_1.ID, test_account_1.ID)
	if err != nil {
		t.Error("неудалось найти ассоциированного пользователя", err.Error())
	}
}
