package models

import (
	"fmt"
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

	// создаем пользователя-владельца для создания тестового аккаунта
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

	// создаем пользователя для тестовой роли
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

	// создаем тестовый аккаунт от имени тестового пользователя-владельца
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

	// добавляем в тестовый аккаунт тестового пользователя
	if err := test_account.AppendUser(&test_user_2); err != nil {
		t.Error("Неудалось добавить пользователя в аккаунт", err.Error())
	}

	// создаем тестовую роль в контексте аккаунта
	test_role_1 := Role{Name:"Test_Role", Tag: "test_tag", Description: "Test crating role for account"}
	if err := test_account.CreateRole(&test_role_1); err != nil {
		t.Error("неудалось создать роль: ", err.Error())
	} else {
		defer func() {
			if err := test_account.DeleteRole(&test_role_1); err != nil {
				t.Error("неудалось удалить роль: ", err.Error())
			}
		}()
	}

	// находим нашего тестового пользователя в представлении AUser
	test_aUser := AccountUser{}
	if err := test_aUser.GetAccountUser(test_user_2.ID, test_account.ID); err != nil {
		t.Error("Неудалось найти aUser: ", test_aUser)
	}

	// устанавливаем роль тестовому пользователю aUser
	if err := test_aUser.SetRole(&test_role_1); err != nil {
		t.Error("Неудалось привязать роль к aUser")
	} else {
		// ставим системную роль, чтобы можно было удалить тестовую роль
		defer func() {
			// 0. Проверим, что можно установить системную роль
			if err := test_aUser.SetRoleManager(); err != nil {
				t.Error("Неудалось отвязать роль от aUser")
			}
		}()
	}

	// # 1. проверяем, что нельзя удалить роль, если к ней привязан хотя бы 1 пользователь
	if err := test_role_1.delete(); err == nil {
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

	test_user_1 := User{
		Username:"user_test_2",
		Email: "mail-test@ratus-dev.ru",
		Name:"РеальноеИмя",
		Surname:"РеальнаяФамилия",
		Patronymic:"РеальноеОтчество",
		Password: "qwerty123#Aa",
	}
	if err := test_user_1.Create(); err != nil {
		t.Error(err.Error())
	} else {
		defer func() {
			if err := test_user_1.Delete(); err != nil {
				t.Error("неудалось удалить пользователя: ", err.Error())
			}
		}()
	}

	test_role_1 := Role{Name:"Test_Role", Tag: "test_tag", Description: "Test crating role for account"}
	if err := test_account.CreateRole(&test_role_1); err != nil {
		t.Error("неудалось создать роль: ", err.Error())
	} else {
		defer func() {
			if err := test_account.DeleteRole(&test_role_1); err != nil {
				t.Error("неудалось удалить роль: ", err.Error())
			}
		}()
	}

	// todo: все что ниже
	// # 1. Проверим можно ли добавить для роли новые разрешения
	/*permissions := []Permission{}
	if err := base.GetDB().Find(&permissions, "code_name = 'PermissionUserListing' OR code_name = 'PermissionUserEditing'").Error; err != nil {
		t.Error("Cant find system permissions: ", err.Error())
	}
	fmt.Println("permissions: ", permissions)*/
	if err := test_role_1.SetPermissions([]int{PermissionStoreEditing,PermissionStoreDeleting,PermissionProductEditing});err != nil {
		t.Error(err)
	}

	permission := Permission{}
	if err := permission.Find(PermissionStoreEditing); err != nil {
		t.Error("Cant find permission code: ", PermissionStoreEditing)
	}
	fmt.Println("permission: ", permission)


	/*p, err := FindPermissions(PermissionStoreEditing)
	if err != nil {
		t.Error("Cant find permission code: ", PermissionStoreEditing)
	}
	fmt.Println("[]Permission{}: ", p)*/

	// # 2. Проверим можно ли удалить у роли разрешения

	// # 3. А теперь узнаем как влияет добавление и удаление на реальные возможности пользователя
	if err := test_account.AppendUser(&test_user_1); err !=nil {
		t.Error("Cant append user to account", test_user_1, test_account)
	}

	aUser := AccountUser{}
	if err := aUser.GetAccountUser(test_user_1.ID, test_account.ID); err != nil {
		t.Error("Cant get account user!", test_user_1, test_account)
	}
	if err := aUser.SetRole(&test_role_1); err != nil {
		t.Error("Cant set new role to test user", test_user_1, test_role_1)
	} else {
		defer func() {
			if err := aUser.SetRoleManager(); err != nil {
				t.Error("Cant set manager to test user", test_user_1, test_role_1, test_account)
			}
		}()
	}






	// todo ...
}

func TestRole_RemovePermissions(t *testing.T) {

}