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

	// создаем пользователя владельца аккаунта
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

	// создаем тестовый аккаунт
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

	// 1. проверим, что нельзя создать думабль аккаунта
	if err := test_account_1.Create(&test_user_1); err == nil {
		t.Error("Удалось повторно создать аккаунт, который был уже создан")
	} else {
		defer func() {
			if err := test_account_1.Delete(); err != nil {
				t.Error("неудалось удалить аккаунт: ", err.Error())
			}
		}()
	}

	// 2. проверяем что аккаунт действительно создан
	temp_account := Account{}
	if base.GetDB().First(&temp_account, test_account_1.ID).RecordNotFound() {
		t.Errorf("Cant find created account: %v", test_account_1.Name)
	}

	// 3. проверим, что в аккаунт был добавлен владелец аккаунта
	if base.GetDB().Model(&test_user_1).Where("account_id = ?", test_account_1.ID).Association("Accounts").Count() != 1 {
		t.Error("Владелец аккаунта не добавлен в список пользователей аккаунта")
	}

	// 4. проверим, что нам не дадут удалить пользователя, если у него есть аккаунты
	err = test_user_1.Delete()
	if err == nil {
		t.Error("Удалось удалить пользователя, при существующем аккаунте")
	}
	if base.GetDB().First(&temp_account, test_account_1.ID).RecordNotFound() {
		t.Errorf("аккаунт был удален, хотя есть ограничение внешнего ключа: %v", test_account_1.Name)
	}

	// 5. проверим, что у пользователя права создателя аккаунта
	owner_role := Role{}
	err = base.GetDB().First(&owner_role, "tag = 'owner'").Error
	if err != nil && !gorm.IsRecordNotFoundError(err){
		t.Error("неудалось найти роль owner", err.Error())
	}
	if gorm.IsRecordNotFoundError(err) {
		t.Skip("Не найдена роль OWNER")
	} else {
		aUser := AccountUser{}
		if err := aUser.GetAccountUser(test_user_1.ID, test_account_1.ID); err != nil {
			t.Error(err.Error())
		}
		// собственно сама проверка роли владельца аккаунта
		if aUser.RoleID != owner_role.ID {
			t.Error("Владелец аккаунта не получил роль создателя", aUser)
		}
	}

	// удаляем пользователя (должен удалиться)
	if err := test_user_1.Delete(); err == nil {
		t.Error("Удалось удалить пользователя, хотя у него был аккаунт.")
	}

}

func TestAccount_ValidateCreate(t *testing.T) {

	// В тестах мы вызываем функцию валидатора, хотя можно вызывать функцию создания аккаунта
	// Тест создания аккаунта в ф-ии теста создания аккаунта

	// создаем пользователя владельца аккаунта
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

	// 1. Убедимся, что простой аккаунт от действительного пользователя создается и тут все Ок
	test_account_1 := Account {	Name:"Account_Test",}
	if myErr := test_account_1.ValidateCreate(&test_user_1); myErr.HasErrors() {
		t.Error("Неудалось пройти валидацию нормальному аккаунту")
	}

	// 2. Убедимся, что аккаунт без имени не будет создан
	test_account_2 := Account {	Name:"",}
	if myErr := test_account_2.ValidateCreate(&test_user_1); !myErr.HasErrors() {
		t.Error("Удалось пройти валидацию аккаунту без имени")
	}

	// 3. Убедимся, что аккаунт с коротким именим Кириллицей не будет создан
	test_account_3 := Account {	Name:"ЦЙ",}
	if myErr := test_account_3.ValidateCreate(&test_user_1); !myErr.HasErrors() {
		t.Error("Удалось пройти валидацию аккаунту с коротким именем")
	}

	// 4. Убедимся, что нельзя создать аккаунт от еще несуществующего пользователя
	test_user_2 := User{
		Username:"user_test_2",
		Email: "test-mail@ratus-dev.ru",
		Name:"РеальноеИмя",
		Surname:"РеальнаяФамилия",
		Patronymic:"РеальноеОтчество",
		Password: "qwerty123#Aa",
	}
	test_account_4 := Account {	Name:"Account_Test",}
	if myErr := test_account_4.ValidateCreate(&test_user_2); !myErr.HasErrors() {
		t.Error("Удалось пройти валидацию аккаунту с несуществующим пользователем")
	}


}

func TestAccount_CreateRole(t *testing.T) {
	// 1. Проверяем НЕ системную роль с привязкой к аккаунту
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

	// создаем тестовую роль в контексте аккаунта (т.е. НЕ системную)
	test_role_1 := Role{Name:"Test_Role", Tag: "test_tag", Description: "Test crating role for account"}
	if err := test_account.CreateRole(&test_role_1, []int{int(PermissionUserAppend)}); err != nil {
		t.Error("неудалось создать роль: ", err.Error())
	} else {
		defer func() {
			if err := test_account.DeleteRole(&test_role_1); err != nil {
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

	// 2. Проверяем системную роль БЕЗ привязки к аккаунту
	test_role_2 := Role{Name:"Test_Role_2", System: true, Tag: "test_tag_2", Description: "Test crating role for account"}
	if err := test_role_2.create([]int{}); err != nil {
		t.Error("неудалось создать роль: ", err.Error())
	} else {
		defer func() {
			if err := test_role_2.delete(); err != nil {
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

func TestAccount_RemoveRole(t *testing.T) {

	// создаем пользователя владельца тестового аккаунта
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

	// создаем тестовый аккаунт, в котором будем создавать тестовые роли
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

	// найдем существующий пермишен, который потребуется для создания роли
	test_permission := Permission{}
	if err := test_permission.Find(PermissionStoreListing); err != nil {
		t.Error("Неудалось найти правило для тестовой роли")
	}

	// создаем тестовую роль в контексте аккаунта (т.е. НЕ системную)
	test_role_1 := Role{Name:"Test_Role", Tag: "test_tag", Description: "Test crating role for account"}
	if err := test_account.CreateRole(&test_role_1, []int{int(test_permission.Code)}); err != nil {
		t.Error("неудалось создать роль: ", err.Error())
	} else {
		defer func() {
			if err := test_account.DeleteRole(&test_role_1); err != nil {
				t.Error("неудалось удалить роль: ", err.Error())
			}
		}()
	}

	// проверим, что функция удаления действительно работает
	if err := test_account.DeleteRole(&test_role_1); err !=nil {
		t.Error("Неудалось удалить тестовую роль", err.Error())
	}

	// проверим, что роль действительно удалилась
	if !base.GetDB().First(&Role{}, test_role_1.ID).RecordNotFound() {
		t.Errorf("Найдена роль, которая должна быть удалена: %v", test_role_1)
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

		if err := test_account.CreateRole(&test_roles[i], []int{}); err != nil {
			t.Error("неудалось создать роль: ", err)
		} else {
			defer func(i int) {
				if err :=  test_account.DeleteRole(&test_roles[i]); err != nil {
					t.Error("неудалось удалить роль: ", err.Error())
				}
			}(i)
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
	if err := test_account.DeleteAllRoles(); err != nil {
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

	// создадим пользователя, от чьего имени будем создавать тестовый аккаунт
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

	// создадим тестовый аккаунт, чтобы в нем создать тестовую роль
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

	// создадим роль в аккаунте, чтобы проверить ее удаление (не совсем отсюда, но может пригодиться)
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



