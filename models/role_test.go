package models

import (
	"github.com/nkokorev/crm-go/database/base"
	"testing"
)

func TestExistRoleTable(t *testing.T) {
	db := base.GetDB()
	if !db.HasTable(&Role{}) {
		tableName := db.NewScope(&Role{}).GetModelStruct().TableName(db)
		t.Errorf("%v table is not exist!",tableName)
	}
}

func TestRole_Delete(t *testing.T) {

	test_user := User{
		Username:"user_test",
		Email: "testmail@ratus-dev.ru",
		Name:"РеальноеИмя",
		Surname:"РеальнаяФамилия",
		Patronymic:"РеальноеОтчество",
		Password: "qwerty123#Aa",
	}
	if err := test_user.Create(); err != nil {
		t.Error(err.Error())
	} else {
		defer func() {
			if err := test_user.Delete(); err != nil {
				t.Error("неудалось удалить пользователя: ", err.Error())
			}
		}()
	}

	test_account := Account {Name:"Account_Test"}
	if err := test_user.CreateAccount(&test_account); err != nil {
		t.Error(err.Error())
	} else {
		defer func() {
			if err := test_account.Delete(); err != nil {
				t.Error("неудалось удалить аккаунт: ", err.Error())
			}
		}()
	}

	test_role_1 := Role{Name:"Test_Role", Tag: "test_tag", AccountID: test_account.ID, Description: "Test crating role for account"}
	if err := test_role_1.Create(); err != nil {
		t.Error("неудалось создать роль: ", err.Error())
	} else {
		defer func() {
			if err := test_role_1.Delete(); err != nil {
				t.Error("неудалось удалить роль: ", err.Error())
			}
		}()
	}

	test_aUser := AccountUser{}
	if err := test_aUser.GetAccountUser(test_user.ID, test_account.ID); err != nil {
		t.Error("Неудалось найти aUser: ", test_aUser)
	}

	if err := test_aUser.SetNewRole(&test_role_1); err != nil {
		t.Error("Неудалось привязать роль к aUser")
	}

	// проверяем, что нельзя удалить роль, если к ней привязан хотя бы 1 пользователь
	if err := test_role_1.Delete(); err == nil {
		t.Error("Удалена роль, хотя к ней привязан пользователь: ", test_user.Name)
	}

	// ставим системную роль, чтобы можно было удалить тестовую роль
	if err := test_aUser.SetAdminRole(); err != nil {
		t.Error("Неудалось отвязать роль от aUser")
	}
}
