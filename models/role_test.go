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

	test_user_owner := User{
		Username:"user_test",
		Email: "testmail@ratus-dev.ru",
		Name:"РеальноеИмя",
		Surname:"РеальнаяФамилия",
		Patronymic:"РеальноеОтчество",
		Password: "qwerty123#Aa",
	}
	if err := test_user_owner.Create(); err != nil {
		t.Error(err.Error())
	} else {
		defer func() {
			if err := test_user_owner.Delete(); err != nil {
				t.Error("неудалось удалить пользователя: ", err.Error())
			}
		}()
	}

	// пользователь для тестовой роли
	test_user_2 := User{
		Username:"user_test_2",
		Email: "mail-test@ratus-dev.ru",
		Name:"РеальноеИмя",
		Surname:"РеальнаяФамилия",
		Patronymic:"РеальноеОтчество",
		Password: "qwerty123#Aa",
	}
	if err := test_user_2.Create(); err != nil {
		t.Error(err.Error())
	} else {
		defer func() {
			if err := test_user_2.Delete(); err != nil {
				t.Error("неудалось удалить пользователя: ", err.Error())
			}
		}()
	}

	test_account := Account {Name:"Account_Test"}
	if err := test_user_owner.CreateAccount(&test_account); err != nil {
		t.Error(err.Error())
	} else {
		defer func() {
			if err := test_account.Delete(); err != nil {
				t.Error("неудалось удалить аккаунт: ", err.Error())
			}
		}()
	}

	if err := test_account.AppendUser(&test_user_2); err != nil {
		t.Error("Неудалось добавить пользователя в аккаунт", err.Error())
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
	if err := test_aUser.GetAccountUser(test_user_2.ID, test_account.ID); err != nil {
		t.Error("Неудалось найти aUser: ", test_aUser)
	}

	if err := test_aUser.SetNewRole(&test_role_1); err != nil {
		t.Error("Неудалось привязать роль к aUser")
	} else {
		// ставим системную роль, чтобы можно было удалить тестовую роль
		defer func() {
			if err := test_aUser.SetAdminRole(); err != nil {
				t.Error("Неудалось отвязать роль от aUser")
			}
		}()
	}

	// проверяем, что нельзя удалить роль, если к ней привязан хотя бы 1 пользователь
	if err := test_role_1.Delete(); err == nil {
		t.Error("Удалена роль, хотя к ней привязан пользователь: ", test_user_2.Name)
	}

}

func TestRole_AppendPermissions(t *testing.T) {

	test_owner_user := User{
		Username:"user_test",
		Email: "testmail@ratus-dev.ru",
		Name:"РеальноеИмя",
		Surname:"РеальнаяФамилия",
		Patronymic:"РеальноеОтчество",
		Password: "qwerty123#Aa",
	}
	if err := test_owner_user.Create(); err != nil {
		t.Error(err.Error())
	} else {
		defer func() {
			if err := test_owner_user.Delete(); err != nil {
				t.Error("неудалось удалить пользователя: ", err.Error())
			}
		}()
	}

	test_account := Account {Name:"Account_Test"}
	if err := test_owner_user.CreateAccount(&test_account); err != nil {
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

	// todo ...
}

func TestRole_RemovePermissions(t *testing.T) {

}