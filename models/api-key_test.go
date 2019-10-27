package models

import (
	"github.com/nkokorev/crm-go/database/base"
	"testing"
)

func TestExistApiKeyTable(t *testing.T) {
	db := base.GetDB()
	if !db.HasTable(&ApiKey{}) {
		tableName := db.NewScope(&ApiKey{}).GetModelStruct().TableName(db)
		t.Errorf("%v table is not exist!",tableName)
	}
}

func TestApiKey_CreateToken(t *testing.T)  {
	test_api_key := ApiKey{}
	if err := test_api_key.createToken();err != nil {
		t.Error("Неудалось создать токен для API Key")
	}
	if len([]rune(test_api_key.Token)) != 32 {
		t.Error("Длина API ключа не равна 32 символам")
	}
}

func TestApiKey_Create(t *testing.T) {

	// создадим пользователя, от чьего имени будем создавать тестовый аккаунт
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

	// создадим тестовый аккаунт, чтобы в нем создать тестовый api-ключ
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

	// находим роль с полным доступом
	test_role_full_access := Role{}
	if err := test_role_full_access.FindRoleByTag("full-access"); err != nil {
		t.Error("Неудалось найти роль с tag: full-access", err.Error())
	}

	// 1. Пробуем создать apiKey
	test_api_key_1 := ApiKey{Name: "TestApiKey_1"}
	if err := test_account.CreateApiKey(&test_api_key_1, &test_role_full_access);err != nil {
		t.Error("Неудалось создать API Key для тестового аккаунта", err.Error())
	}

	// убеждаемся, что новый api-ключ создан
	if test_api_key_1.ID == 0 {
		t.Errorf("ApiKey ID == 0, expected > 0")
	}

	// 2. Проверим назначение ролей
	if err := test_api_key_1.SetRoleFullAccess(); err != nil {
		t.Error("Неудалось назначить роль FullAssets для API Key", err.Error())
	}
	if err := test_api_key_1.SetRoleSiteAccess(); err != nil {
		t.Error("Неудалось назначить роль SiteAccess для API Key", err.Error())
	}
	if err := test_api_key_1.SetRoleReadAccess(); err != nil {
		t.Error("Неудалось назначить роль ReadAccess для API Key", err.Error())
	}

	// 2. Проверяем, что после удаления аккаунта, ключик удаляется (связанность внешним ключем с аккаунтом)
	if err := test_account.Delete(); err !=nil {
		t.Error("Неудалось удалить аккаунт", err.Error())
	}
	if !base.GetDB().First(&ApiKey{}, test_api_key_1.ID).RecordNotFound() {
		t.Error("Найден ключ, которого быть не должно :)", test_api_key_1)
	}
}

func TestApiKey_Delete(t *testing.T) {

	// создадим пользователя, от чьего имени будем создавать тестовый аккаунт
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

	// создадим тестовый аккаунт, чтобы в нем создать тестовый api-ключ
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

	// находим системную роль
	test_role_full_access := Role{}
	if err := test_role_full_access.FindRoleByTag("site-access"); err != nil {
		t.Error("Неудалось найти роль с tag: full-access", err.Error())
	}

	// 1. Пробуем создать apiKey
	test_api_key_1 := ApiKey{Name: "TestApiKey_1"}
	if err := test_account.CreateApiKey(&test_api_key_1, &test_role_full_access);err != nil {
		t.Error("Неудалось создать API Key для тестового аккаунта", err.Error())
	}

	// убеждаемся, что новый api-ключ создан
	if test_api_key_1.ID == 0 {
		t.Errorf("ApiKey ID == 0, expected > 0")
	}

	// 2. Проверяем удаление ключа
	token := test_api_key_1.Token // запоминаем токен
	if err := test_account.DeleteApiKey(&test_api_key_1); err !=nil {
		t.Error("Неудалось удалить аккаунт", err.Error())
	}
	// убеждаемся, что ключ удален
	if !base.GetDB().First(&ApiKey{}, "token = ?", token).RecordNotFound() {
		t.Error("Найден ключ, которого быть не должно :)", test_api_key_1)
	}

}

func TestApiKey_SetRole(t *testing.T) {

	// создадим пользователя, от чьего имени будем создавать тестовый аккаунт
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

	// создадим тестовый аккаунт, чтобы в нем создать тестовый api-ключ
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

	// находим роль с полным доступом, чтобы протестировать с ним ключ
	test_role_full_access := Role{}
	if err := test_role_full_access.FindRoleByTag("full-access"); err != nil {
		t.Error("Неудалось найти роль с tag: full-access", err.Error())
	}

	// создаем тестовую роль в контексте аккаунта, чтобы затестить ключ
	test_role_1 := Role{Name:"Test_Role", Tag: "test_tag", Type: "api", Description: "Test crating role for account"}
	if err := test_account.CreateRole(&test_role_1); err != nil {
		t.Error("неудалось создать роль: ", err.Error())
	} else {
		defer func() {
			if err := test_account.DeleteRole(&test_role_1); err != nil {
				t.Error("неудалось удалить роль: ", err.Error())
			}
		}()
	}

	// 1. Пробуем создать apiKey с системной ролью
	test_api_key_1 := ApiKey{Name: "TestApiKey_1"}
	if err := test_account.CreateApiKey(&test_api_key_1, &test_role_full_access);err != nil {
		t.Error("Неудалось создать API Key для тестового аккаунта", err.Error())
	}
	// убеждаемся, что новый api-ключ создан
	if test_api_key_1.ID == 0 {
		t.Errorf("ApiKey ID == 0, expected > 0")
	}

	// 2. Пробуем создать apiKey с нашей ролью в контексте аккаунта
	test_api_key_2 := ApiKey{Name: "TestApiKey_2"}
	if err := test_account.CreateApiKey(&test_api_key_2, &test_role_1);err != nil {
		t.Error("Неудалось создать API Key для тестового аккаунта", err.Error())
	}
	// убеждаемся, что новый api-ключ создан
	if test_api_key_2.ID == 0 {
		t.Errorf("ApiKey ID == 0, expected > 0")
	}
}
