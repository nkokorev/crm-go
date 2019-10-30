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
	test_user_owner := User{
		Username:"test_user_owner",
		Email: "testmail@ratus-dev.ru",
		Name:"РеальноеИмя",
		Surname:"РеальнаяФамилия",
		Patronymic:"РеальноеОтчество",
		Password: "qwerty123#Aa",
	}
	err := test_user_owner.Create()
	if err != nil {
		t.Error(err.Error())
	} else {
		defer func() {
			if err := test_user_owner.Delete(); err != nil {
				t.Error("неудалось удалить пользователя: ", err.Error())
			}
		}()
	}

	test_account_1 := Account {	Name:"Account_Test",}
	err = test_user_owner.CreateAccount(&test_account_1)
	if err != nil {
		t.Error(err.Error())
	} else {
		defer func() {
			if err := test_account_1.Delete(); err != nil {
				t.Error("неудалось удалить аккаунт: ", err.Error())
			}
		}()
	}

	test_acc_user_1 := AccountUser{}
	err = test_acc_user_1.GetAccountUser(test_user_owner.ID, test_account_1.ID)
	if err != nil {
		t.Error("неудалось найти ассоциированного пользователя", err.Error())
	}
}

func TestAccountUser_SetNewRole(t *testing.T) {

	test_owner_user := User{
		Username:"test_user_owner",
		Email: "testmail@ratus-dev.ru",
		Name:"РеальноеИмя",
		Surname:"РеальнаяФамилия",
		Patronymic:"РеальноеОтчество",
		Password: "qwerty123#Aa",
	}
	if err := test_owner_user.Create(); err != nil {
		t.Error(err.Error())
		return
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
		return
	} else {
		defer func() {
			if err := test_account.Delete(); err != nil {
				t.Error("неудалось удалить аккаунт: ", err.Error())
			}
		}()
	}

	test_role_1 := Role{Name:"Test_Role", Tag: "test_tag", AccountID: test_account.ID, Description: "Test crating role for account"}
	if err := test_account.CreateRole(&test_role_1, []int{}); err != nil {
		t.Error("неудалось создать роль: ", err.Error())
		return
	} else {
		defer func() {
			if err := test_account.DeleteRole(&test_role_1); err != nil {
				t.Error("неудалось удалить роль: ", err.Error())
			}
		}()
	}

	test_owner_account_user := AccountUser{}
	if err := test_owner_account_user.GetAccountUser(test_owner_user.ID, test_account.ID); err != nil {
		t.Error("неудалось найти ассоциированного пользователя", err.Error())
	}

	// 1. Убеждаемся, что создавший аккаунт пользователь имеет роль владельца
	temp_role_1 := Role{}
	if err := base.GetDB().Model(&test_owner_account_user).Related(&temp_role_1).Error; err != nil {
		t.Error("Неудалось найти роль владельца аккаунта", err.Error())
	}
	if temp_role_1.Tag != "owner" {
		t.Error("Владелец аккаунта не получил роль владельца аккаунта", temp_role_1, test_account, test_owner_user)
	}

	// 2. Убеждаемся, что нельзя назначить новую роль владельцу аккаунта
	if err := test_owner_account_user.SetRoleManager(); err == nil {
		t.Error("Перенезначение роли для роли owner! : ", test_owner_account_user, test_role_1, test_account)
	}

	// 3. Проверяем назначение роли для нового пользователя
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
		return
	} else {
		defer func() {
			if err := test_user_2.Delete(); err != nil {
				t.Error("неудалось удалить пользователя: ", err.Error())
			}
		}()
	}
	if _,err := test_account.AppendUser(&test_user_2); err != nil {
		t.Error("Невышло добавить пользователя в аккаунт", test_account, test_user_2)
	}
	test_account_user := AccountUser{}
	if err := test_account_user.GetAccountUser(test_user_2.ID, test_account.ID); err != nil {
		t.Error("Неудалось найти ассоциированного пользователя")
	}
	if err := test_account_user.SetRole(&test_role_1); err != nil {
		t.Error("Неудалось назначить новую роль пользователю, не являющимся владельцем аккаунта")
	}

}

// ниже функции как один проверка назначения ролей
func TestAccountUser_SetRoleOwner(t *testing.T) {
	// todo ... написать простой тест
}
func TestAccountUser_SetRoleAdmin(t *testing.T) {
	// todo ... написать простой тест
}
func TestAccountUser_SetRoleManager(t *testing.T) {
	// todo ... написать простой тест
}
func TestAccountUser_SetRoleMarketer(t *testing.T) {
	// todo ... написать простой тест
}
func TestAccountUser_SetRoleAuthor(t *testing.T) {
	// todo ... написать простой тест
}
func TestAccountUser_SetRoleViewer(t *testing.T) {
	// todo ... написать простой тест
}


func TestAccountUser_CheckPermission(t *testing.T) {
	// в отличии от (*Role) HasPermissions() bool тут надо проверить среди пользователей с разными ролями

	// создаем владельца аккаунта
	test_owner_user := User{
		Username:"user_owner",
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

	// создаем тестовый аккаунт, от имени владельца аккаунта
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

	// создаем обычного пользователя, которого потом добавим в аккаунт и будем тестировать на нем различные права
	test_user_1 := User{
		Username:"test_user_1",
		Email: "test-mail@ratus-dev.ru",
		Name:"РеальноеИмя",
		Surname:"РеальнаяФамилия",
		Patronymic:"РеальноеОтчество",
		Password: "qwerty123#Aa",
	}
	if err := test_user_1.Create(); err != nil {
		t.Error(err.Error())
		return
	} else {
		defer func() {
			if err := test_user_1.Delete(); err != nil {
				t.Error("неудалось удалить пользователя: ", err.Error())
			}
		}()
	}

	// добавляем подопытного в новенький аккаунт
	test_aUser_1 := &AccountUser{}
	test_aUser_1, err := test_account.AppendUser(&test_user_1)
	if err != nil {
		t.Error("Неудалось добавить пользователя в тестовый аккаунт")
	}

	// получаем aUserOwner
	test_aUserOwner := AccountUser{}
	if err := test_aUserOwner.GetAccountUser(test_owner_user.ID, test_account.ID); err != nil {
		t.Error("Не вышло найти test_aUserOwner")
		return
	}
	// получаем aUser_1
	/*test_aUser_1 := AccountUser{}
	if err := test_aUser_1.GetAccountUser(test_user_1.ID, test_account.ID); err != nil {
		t.Error("Не вышло найти test_aUser_1")
		return
	}*/

	// 1. Проверим права у владельца аккаунта
	if ! test_aUserOwner.CheckPermission(PermissionBillingManagement) {
		t.Error("У владельца нет крутых ролей!")
	}
	if ! test_aUserOwner.CheckPermission(PermissionAPIManagement) {
		t.Error("У владельца нет крутых ролей!")
	}

	// 2. Проверим отсутствие прав у безправного юзера
	if test_aUser_1.CheckPermission(PermissionBillingManagement) {
		t.Error("У простого пользователя нашлись крутые права!")
	}
	if test_aUser_1.CheckPermission(PermissionAPIManagement) {
		t.Error("У простого пользователя нашлись крутые права!")
	}
	if test_aUser_1.CheckPermission(PermissionProductListing) {
		t.Error("У простого пользователя нашлись права!")
	}

	// 3. Добавим роль пользователю и убедимся, что она чекается
	if err := test_aUser_1.SetRoleAdmin();err != nil {
		t.Error("Неудалось назначить роль администратора пользователю")
	}
	if !test_aUser_1.CheckPermission(PermissionProductListing) {
		t.Error("Пользователю дали админа, а он ничего не может!")
	}

	// огонь, мы молодцы!
}

