package models

import (
	"github.com/jinzhu/gorm"
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
		t.Error("Неудалось создать пользователя: ", err.Error())
	} else {
		defer func() {
			if err := test_user_1.Delete(); err != nil {
				t.Error("неудалось удалить пользователя: ", err.Error())
			}
		}()
	}

	test_account_1 := Account {	Name:"Account_Test",}
	err = test_account_1.Create(&test_user_1)
	if err != nil {
		t.Error("Неудалось создать аккаунт: ", err.Error())
	} else {
		defer func() {
			if err := test_account_1.Delete(); err != nil {
				t.Error("неудалось удалить аккаунт: ", err.Error())
			}
		}()
	}

	// #1. проверяем что аккаунт действительно создан
	temp_account := Account{}
	if base.GetDB().First(&temp_account, test_account_1.ID).RecordNotFound() {
		t.Errorf("Cant find created account: %v", test_account_1.Name)
	}

	// #2. проверим, что в аккаунт был добавлен владелец аккаунта
	if base.GetDB().Model(&test_user_1).Where("account_id = ?", test_account_1.ID).Association("Accounts").Count() != 1 {
		t.Error("Владелец аккаунта не добавлен в список пользователей аккаунта")
	}

	// #3. проверим, что нам не дадут удалить пользователя, если у него есть аккаунты
	err = test_user_1.Delete()
	if err == nil {
		t.Error("Удалось удалить пользователя, при существующем аккаунте")
	}
	if base.GetDB().First(&temp_account, test_account_1.ID).RecordNotFound() {
		t.Errorf("аккаунт был удален, хотя есть ограничение внешнего ключа: %v", test_account_1.Name)
	}

	// #4. проверим, что у пользователя права создателя аккаунта
	owner_role := Role{}
	err = base.GetDB().First(&owner_role, "tag = 'owner'").Error
	if err != nil && !gorm.IsRecordNotFoundError(err){
		t.Error("неудалось создать роль", err.Error())
	}
	if gorm.IsRecordNotFoundError(err) {
		t.Skip("Не найдена роль OWNER")
	} else {
		aUser := AccountUser{}
		if err := aUser.GetAccountUser(test_user_1.ID, test_account_1.ID); err != nil {
			t.Error(err.Error())
		}
		if aUser.RoleID != owner_role.ID {
			t.Error("Владелец аккаунта не получил роль создателя", aUser)
		}
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

func TestAccount_CreateRole(t *testing.T) {
	// #1. Проверяем НЕ системную роль с привязкой к аккаунту
	// (только такая и должна быть, т.к. системные роли без привязки к аккаунту идут)
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

	test_role_1 := Role{Name:"Test_Role", Tag: "test_tag_1", AccountID: test_account.ID, Description: "Test crating role for account"}
	if err := test_role_1.Create(); err != nil {
		t.Error("неудалось создать роль: ", err.Error())
	} else {
		defer func() {
			if err := test_role_1.Delete(); err != nil {
				t.Error("неудалось удалить роль: ", err.Error())
			}
		}()
	}

	// проверяем, что роль действительно создана
	temp_role_1 := Role{}
	if base.GetDB().First(&temp_role_1, test_role_1.ID).RecordNotFound() {
		t.Errorf("Cant find created account: %v", test_role_1.Name)
	}

	// проверяем, что новая роль НЕ системная
	if test_role_1.System {
		t.Errorf("созданная роль оказалась системной, хотя не должна быть таковой: %v", test_role_1.Name)
	}

	// #2. Проверяем системную роль БЕЗ привязки к аккаунту
	test_role_2 := Role{Name:"Test_Role", System: true, Tag: "test_tag_2", Description: "Test crating role for account"}
	if err := test_role_2.Create(); err != nil {
		t.Error("неудалось создать роль: ", err.Error())
	} else {
		defer func() {
			if err := test_role_2.Delete(); err != nil {
				t.Error("неудалось удалить роль: ", err.Error())
			}
		}()
	}

	// проверяем, что роль действительно создана
	temp_role_2 := Role{}
	if base.GetDB().First(&temp_role_2, test_role_2.ID).RecordNotFound() {
		t.Errorf("Cant find created account: %v", test_role_2.Name)
	}

	// проверяем, что новая роль СИСТЕМНАЯ
	if !test_role_2.System {
		t.Errorf("созданная роль оказалась НЕ системной, хотя не должна быть таковой: %v", test_role_2.Name)
	}

}

func TestAccount_RemoveAllRoles(t *testing.T) {

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

	// Эти роли должны удалиться
	test_roles := []Role{
		{Name:"Test_Role_1", AccountID: test_account.ID, Tag: "test_tag_1", Description: "Test crating role for account"},
		{Name:"Test_Role_2", AccountID: test_account.ID, Tag: "test_tag_2", Description: "Test crating role for account"},
		{Name:"Test_Role_3", AccountID: test_account.ID, Tag: "test_tag_3", Description: "Test crating role for account"},
		}

	for i, _ := range test_roles {
		err := test_roles[i].Create();
		if err != nil || test_roles[i].ID == 0 {
			t.Error("неудалось создать роль: ", err)
		} else {
			/*defer func(i int) {
				if err :=  test_roles[i].Delete(); err != nil {
					t.Error("неудалось удалить роль: ", err.Error())
				}
			}(i)*/
		}
	}

	// проверяем, что данные есть
	if base.GetDB().Model(&test_account).Association("Roles").Count() != 3 {
		t.Errorf("Неудалось создать тестовые роли для аккаунта или не совпадает число созданных ролей: %v", test_account.Name)
	}

	// проверим наличие системных ролей
	countSystemRoles := 0
	base.GetDB().Model(&Role{}).Where("system = TRUE").Count(&countSystemRoles)
	if countSystemRoles < 1 {
		t.Error("В базе отсутствуют системные роли.")
	}

	// #1. Проверим удаление привязанных к аккаунту ролей
	if err := test_account.RemoveAllRoles(); err != nil {
		t.Errorf("Неудалось удалить тестовые роли для аккаунта : %v", test_account.Name)
	}

	// проверяем, что ролей аккаунта больше нет в системе
	if base.GetDB().Model(&test_account).Association("Roles").Count() != 0 {
		t.Errorf("Неудалось удалить тестовые роли для аккаунта или число ролей не равно 0: %v", test_account.Name)
	}

	// #2. Проверяем, что системные роли на месте (!)
	temp_countSystemRoles := 0
	base.GetDB().Model(&Role{}).Where("system = TRUE").Count(&temp_countSystemRoles)
	if countSystemRoles != temp_countSystemRoles {
		t.Fatal("При удалении ролей аккаунта были удалены системные роли!!!")
	}
}

func TestAccount_Delete(t *testing.T) {

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

	// проверяем, что роль действительно создана
	temp_account := Account{}
	if base.GetDB().First(&temp_account, test_account.ID).RecordNotFound() {
		t.Errorf("Cant find created account: %v", test_account.Name)
	}

	// #1. Проверим, удаление аккаунта
	if err := test_account.Delete();err != nil {
		t.Error("неудалось удалить аккаунт: ", err.Error())
	}

	// проверям, что аккаунт удалился
	if !base.GetDB().First(&Account{}, "hash_id = ?",test_account.HashID).RecordNotFound() {
		t.Errorf("найден аккаунт, который должен быть удален hash_id: %v", test_account.HashID)
	}

	// #2. Проверим, что созданная роль тоже удалилась
	if !base.GetDB().First(&Role{}, "hash_id = ?",test_role_1.HashID).RecordNotFound() {
		t.Errorf("найдена роль, которая должна быть удалена hash_id: %v", test_role_1.HashID)
	}

	// #3. Проверим, что пользователь (owner) остался на месте
	if base.GetDB().First(&User{}, "hash_id = ?",test_user.HashID).RecordNotFound() {
		t.Errorf("удален пользователь после удаления его аккаунта, user hash_id: %v", test_user.HashID)
	}

}
