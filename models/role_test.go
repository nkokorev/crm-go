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

func TestExistRoleData(t *testing.T) {
	if base.GetDB().Find(&Role{}).RecordNotFound() {
		t.Error("В базе отсутствуют базовые роли")
	}
}

func TestRole_Delete(t *testing.T) {

	// создаем пользователя-владельца для создания тестового аккаунта
	test_user_owner := User{
		Username:"test_user_owner",
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
	if _,err := test_account.AppendUser(&test_user_2); err != nil {
		t.Error("Неудалось добавить пользователя в аккаунт", err.Error())
	}

	// создаем тестовую роль в контексте аккаунта
	test_role_1 := Role{Name:"Test_Role", Tag: "test_tag", Description: "Test crating role for account"}
	if err := test_account.CreateRole(&test_role_1, []int{}); err != nil {
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
		Username:"test_user_owner",
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

	// создаем роль с 0 разрешений.
	test_role_1 := Role{Name:"Test_Role", Tag: "test_tag", Description: "Test crating role for account"}
	if err := test_account.CreateRole(&test_role_1, []int{}); err != nil {
		t.Error("неудалось создать роль: ", err.Error())
	} else {
		defer func() {
			if err := test_account.DeleteRole(&test_role_1); err != nil {
				t.Error("неудалось удалить роль: ", err.Error())
			}
		}()
	}

	// 1. Проверим, что у новой роли нет прав
	if base.GetDB().Model(&test_role_1).Association("Permissions").Count() > 0 {
		t.Error("У новой роли есть права, хотя быть их не должно")
	}

	// 2. Проверим назначение правила
	if err := test_role_1.AppendPermissions([]int{PermissionUserAppend}); err != nil {
		t.Error("Неудалось назначить права для тестовой роли")
	}

	// 3. Проверим, что правило назначилось для роли-1
	temp_permission := Permission{}
	if err := temp_permission.Find(PermissionUserAppend); err != nil {
		t.Error("Неудалось найти правило PermissionStoreListing")
	}
	// должен найти одно(1) соответствие
	if base.GetDB().Model(&test_role_1).Where("permission_id = ?", temp_permission.ID).Association("Permissions").Count() != 1 {
		t.Error("Неудалось найти назначенное правило для роли")
	}

}

func TestRole_RemovePermissions(t *testing.T) {
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

	// создаем роль, с PermissionStoreListing.
	test_role_1 := Role{Name:"Test_Role", Tag: "test_tag", Description: "Test crating role for account"}
	if err := test_account.CreateRole(&test_role_1, []int{PermissionStoreListing}); err != nil {
		t.Error("неудалось создать роль: ", err.Error())
	} else {
		defer func() {
			if err := test_account.DeleteRole(&test_role_1); err != nil {
				t.Error("неудалось удалить роль: ", err.Error())
			}
		}()
	}

	// 1. Проверим, что правило назначилось для роли-1 (можем проверить через check permissions... но не будем)
	temp_permission := Permission{}
	if err := temp_permission.Find(PermissionStoreListing); err != nil {
		t.Error("Неудалось найти правило PermissionStoreListing")
	}
	// должен найти одно(1) соответствие
	if base.GetDB().Model(&test_role_1).Where("permission_id = ?", temp_permission.ID).Association("Permissions").Count() != 1 {
		t.Error("Неудалось найти назначенное правило для роли")
	}

	// теперь проверим, что правило удалилось
	if err := test_role_1.RemovePermissions([]int{PermissionStoreListing}); err != nil {
		t.Error("Неудалось удалить права у роли")
	}
	// проверим, что права удалились
	if base.GetDB().Model(&test_role_1).Where("permission_id = ?", temp_permission.ID).Association("Permissions").Count() != 0 {
		t.Error("Нашлось правило для роли, которое мы убрали у нее")
	}
}

func TestRole_hasPermission(t *testing.T)  {

	// создаем владельца аккаунта
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

	// создаем роль, с test_permission.
	test_role_1 := Role{Name:"Test_Role", Tag: "test_tag", Description: "Test crating role for account"}
	if err := test_account.CreateRole(&test_role_1, []int{PermissionStoreListing}); err != nil {
		t.Error("неудалось создать роль: ", err.Error())
	} else {
		defer func() {
			if err := test_account.DeleteRole(&test_role_1); err != nil {
				t.Error("неудалось удалить роль: ", err.Error())
			}
		}()
	}

	// 1. Проверяем, что роль имеет нужный нам пермишен
	if !test_role_1.hasPermission(PermissionStoreListing) {
		t.Error("Тестовая роль не получила пермишен PermissionStoreListing код: ", PermissionStoreListing)
	}
	// 2. Проверим, что роль НЕ имеет второй пермишен, который мы не назначали роли
	if test_role_1.hasPermission(PermissionUserAppend) {
		t.Error("Тестовая роль получила пермишен PermissionStoreListing код: ", PermissionUserAppend)
	}

	// 3. Проверим, что добавленная роль, теперь проверяется
	if err := test_role_1.AppendPermissions([]int{PermissionUserAppend}); err != nil {
		t.Error("Неудалось назначить новое правило для роли")
	}
	// Теперь правило [PermissionUserAppend] должно присутствовать
	if ! test_role_1.hasPermission(PermissionUserAppend) {
		t.Error("Тестовая роль НЕ получила пермишен PermissionStoreListing код: ", PermissionUserAppend)
	}

	// 4. Проверим удаление прав
	if err := test_role_1.RemovePermissions([]int{PermissionStoreListing}); err != nil {
		t.Error("Неудалось удалить правило для роли")
	}
	if test_role_1.hasPermission(PermissionStoreListing) {
		t.Error("Тестовая роль имеет пермишен PermissionStoreListing, хотя он был удален, код: ", PermissionStoreListing)
	}

}