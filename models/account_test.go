package models

import (
	"github.com/nkokorev/crm-go/utils"
	"strings"
	"testing"
)

func TestAccount_ValidateInputs(t *testing.T) {

	account := Account{Name:""}
	if err := account.ValidateInputs(); err == nil {
		t.Fatal("Validate account without name")
	}

	account.Name = utils.RandStringBytes(100)
	if err := account.ValidateInputs(); err == nil {
		t.Fatal("Validate account with very long name")
	}

	account.Name = utils.RandStringBytes(10)
	if err := account.ValidateInputs(); err != nil {
		t.Fatal("No validate account with shot name")
	}

	account.Website = utils.RandStringBytes(256)
	if err := account.ValidateInputs(); err == nil {
		t.Fatal("Validate account with very long website name")
	}

	account.Website = utils.RandStringBytes(50)
	if err := account.ValidateInputs(); err != nil {
		t.Fatal("No Validate account with norm website name")
	}

	account.Type = utils.RandStringBytes(256)
	if err := account.ValidateInputs(); err == nil {
		t.Fatal("Validate account with very long type")
	}

	account.Type = utils.RandStringBytes(50)
	if err := account.ValidateInputs(); err != nil {
		t.Fatal("No Validate account with norm type")
	}

}

func TestAccount_createAccount(t *testing.T) {

	// 1. Аккаунт не должен создаваться без вводных данных
	testAccount, err := Account{}.create()
	if err == nil && testAccount != nil {
		defer testAccount.HardDelete()
		t.Fatal("Created account, but name is null")
	}

	outAccount, err := Account{Name: "Test account"}.create(  )
	if err != nil || outAccount == nil {
		t.Fatal("Cant create account without name")
	}

	defer outAccount.HardDelete()

}

func TestGetAccount(t *testing.T) {
	// создаем тестовый аккаунт и потом его находим
	account, err := Account{Name:"Test Get account: " + utils.RandStringBytes(10)}.create()
	if err != nil {
		t.Fatalf("Неудалось создать тестовый аккаунт: %v", err)
	}
	defer account.HardDelete()

	fAccount, err := GetAccount(account.ID)
	if err != nil {
		t.Fatal("Ошибка поиска аккаунта")
	}
	if fAccount == nil {
		t.Fatal("Неудалось найти вновь созданный аккаунт")
	}
	if fAccount.Name != account.Name {
		t.Fatalf("Поиск аккаунта дал не верный результат")
	}
}

func TestGetMainAccount(t *testing.T) {
	account, err := GetMainAccount()
	if err != nil || account.ID != 1 || account.Name != "RatusCRM" {
		t.Fatalf("Cant find main account: %v", err)
	}
}

func TestAccount_CreateApiKey(t *testing.T) {
	account, _ := Account{Name:"Test account for API Key"}.create()
	defer account.HardDelete()

	key, err := account.CreateApiKey(ApiKey{Name:"Api key for Test"})
	if err != nil {
		t.Fatalf("Неудалось создать api-ключ для аккаунта: %v", err)
	}

	defer key.delete()
}

func TestAccount_DeleteApiKey(t *testing.T) {
	account, _ := Account{Name:"Test account for API Key"}.create()
	defer account.HardDelete()

	key, err := account.CreateApiKey(ApiKey{Name:"Api key for Test"})
	if err != nil {
		t.Fatalf("Неудалось создать api-ключ для аккаунта: %v", err)
	}

	// убеждаем, что сначала он его находит
	sKey, err := account.GetApiKey(key.Token)
	if err != nil || sKey == nil {
		t.Fatal("Ошибка с поиском ApiKey - он должен был найтись")
	}

	// убеждаем, что нельзя удалить ключ из-под другого аккаунта
	account2, _ := Account{Name:"Test account for API Key 2"}.create()
	defer account2.HardDelete()

	err = account2.DeleteApiKey(key.Token)
	if err == nil {
		t.Fatal("удалось удалить ApiKey из-под несвязанного аккаунта")
	}

	// а вот теперь должно удалиться
	err = account.DeleteApiKey(key.Token)
	if err != nil {
		t.Fatalf("Неудалось удалить ApiKey: %v", err)
	}

	// убеждаемся, что после удаления нашего ключика нет
	_, err = account.GetApiKey(key.Token)
	if err == nil {
		t.Fatal("Найден apiKey, который был удален")

	}
}

func TestAccount_GetApiKey(t *testing.T) {

	account, _ := Account{Name:"Test account for API Key"}.create()
	defer account.HardDelete()

	account2, _ := Account{Name:"Test account for API Key 2"}.create()
	defer account2.HardDelete()

	key, _ := account.CreateApiKey(ApiKey{Name:"Api key for Test"})
	defer account.DeleteApiKey(key.Token)

	// убеждаем, что нельзя получить ключ из-под другого аккаунта
	_, err := account2.GetApiKey(key.Token)
	if err == nil {
		t.Fatal("удалось получить ApiKey из-под несвязанного аккаунта")
	}
}

func TestAccount_UpdateApiKey(t *testing.T) {
	account, _ := Account{Name:"Test account for API Key"}.create()
	defer account.HardDelete()

	key, _ := account.CreateApiKey(ApiKey{Name:"Api key for Test: " + utils.RandStringBytes(5)})
	defer account.DeleteApiKey(key.Token)

	// Проверим, что новые данные сохраняются и не сохраняются лишние
	token := key.Token
	key.Name = utils.RandStringBytes(10) // должно сработать
	key.Enabled = !key.Enabled // должно сработать
	key.AccountID = key.AccountID + 1 // НЕ должно сработать
	key.Token = utils.RandStringBytes(10) // НЕ должно сработать

	if err := key.update(*key); err !=nil {
		t.Fatalf("Неудалось обновить ApiKey")
	}

	sKey, err := account.GetApiKey(token)
	if err != nil {
		t.Fatal("Неудалось найти ApiKey после update")
	}

	if sKey.Token == key.Token {
		t.Fatal("Удалось обновлением изменить token у ApiKey")
	}
	if sKey.Enabled != key.Enabled {
		t.Fatal("Удалось обновлением изменить Enabled у ApiKey")
	}
	if sKey.AccountID != account.ID {
		t.Fatal("Удалось обновлением изменить AccountID у ApiKey")
	}
}

func TestAccount_CreateUser(t *testing.T) {

	// аккаунт с регистрацией по имени пользователя
	account, err := Account{Name:"Test account CreateUser Username"}.create()
	if err != nil {
		t.Fatalf("Неудалось создать тестовый аккаунт: %v", err)
	}
	defer account.HardDelete()

	// todo дописать список тестов
	testList := []struct {
		account *Account
		user User
		expected bool
		description string
		}{
			{account, User{Username:"", Email:"", Phone:""}, false, "Хотя бы один из определяющих полей должен быть"},
			{account, User{Username:"TestUser 1", Email:"adnsls!@.ru"}, false, "Не корректный email"},
			{account, User{Username:"TestUser 1", Phone:"5456a45355"}, false, "Не корректный телефон"},
		}


	for i, v := range testList {
		user, err := v.account.CreateUser(v.user)

		if v.expected == false && err == nil {
			t.Fatalf("Создан пользователь, которого быть не должно : [%v] user: %v", i, user)
		}

		if v.expected == true && err != nil {
			t.Fatalf("Неудалось создать пользователя, который должен быть создан: [%v] user: %v", i, user)
		}

		// удаляем созданного пользователя
		if err == nil && user != nil {
			user.hardDelete()
		}

	}
}

func TestAccount_GetUserById(t *testing.T) {
	account, err := Account{Name:"TestAccount_GetUserById"}.create()
	if err != nil {
		t.Fatalf("Неудалось создать тестовый аккаунт: %v", err)
	}
	defer account.HardDelete()

	user, err := account.CreateUser(User{Username:"TestUser"})
	if err!=nil {
		t.Fatalf("Неудалось создать пользователя %v", err)
	}
	defer user.hardDelete()

	userF, err := account.GetUserById(user.ID)
	if err != nil {
		t.Fatalf("Неудалось найти пользователя, %v", err)
	}

	if userF.ID != user.ID || userF.ID == 0 {
		t.Fatalf("Ошибка: пользователь найден не правильно!")
	}

}

func TestAccount_GetUserByUsername(t *testing.T) {
	account, err := Account{Name:"TestAccount_GetUserById"}.create()
	if err != nil {
		t.Fatalf("Неудалось создать тестовый аккаунт: %v", err)
	}
	defer account.HardDelete()

	user, err := account.CreateUser(User{Username: utils.RandStringBytes(10)})
	if err!=nil {
		t.Fatalf("Неудалось создать пользователя %v", err)
	}
	defer user.hardDelete()

	fUser, err := account.GetUserByUsername(user.Username)
	if err != nil {
		t.Fatalf("Неудалось найти пользователя, %v", err)
	}

	if fUser.ID != user.ID || fUser.ID == 0 || fUser.Username != user.Username {
		t.Fatalf("Ошибка: пользователь найден не правильно!")
	}

}

func TestAccount_GetUserByEmail(t *testing.T) {
	account, err := Account{Name:"TestAccount_GetUserById"}.create()
	if err != nil {
		t.Fatalf("Неудалось создать тестовый аккаунт: %v", err)
	}
	defer account.HardDelete()

	user, err := account.CreateUser(User{Email: strings.ToLower(utils.RandStringBytes(5)) + "@rus-marketing.com"})
	if err!=nil {
		t.Fatalf("Неудалось создать пользователя %v", err)
	}
	defer user.hardDelete()

	fUser, err := account.GetUserByEmail(user.Email)
	if err != nil {
		t.Fatalf("Неудалось найти пользователя, %v", err)
	}

	if fUser.ID != user.ID || fUser.ID == 0 || fUser.Email != user.Email {
		t.Fatalf("Ошибка: пользователь найден не правильно!")
	}

}

func TestAccount_GetUserByPhone(t *testing.T) {
	account, err := Account{Name:"TestAccount_GetUserByPhone"}.create()
	if err != nil {
		t.Fatalf("Неудалось создать тестовый аккаунт: %v", err)
	}
	defer account.HardDelete()

	user, err := account.CreateUser(User{Phone: "88251001212"})
	if err!=nil {
		t.Fatalf("Неудалось создать пользователя %v", err)
	}
	defer user.hardDelete()

	fUser, err := account.GetUserByPhone(user.Phone, "")
	if err != nil {
		t.Fatalf("Неудалось найти пользователя, %v", err)
	}

	if fUser.ID != user.ID || fUser.ID == 0 || fUser.Phone != user.Phone {
		t.Fatalf("Ошибка: пользователь найден не правильно!")
	}

}




func BenchmarkGetAccountByHash(b *testing.B) {
	// создаем много аккаунтов
	return
	/*runAccounts := 50000
	for i:=0;i < runAccounts;i++ {
		_, err := Account{Name:"TestAccount"}.create()
		if err != nil {
			b.Fatalf("Неудалось создать акаунт, %v", err)
		}
		//defer account.HardDelete()
	}
	return*/
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		acc, err := GetAccountByHash("arfjdlafkl")
		if err != nil || acc == nil {
			b.Fatalf("Неудалось найти аккаунт: %v", err)
		}
	}

	b.StopTimer()

}

func BenchmarkGetAccount(b *testing.B) {

	return

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		acc, err := GetAccount(41047)
		if err != nil || acc == nil {
			b.Fatalf("Неудалось найти аккаунт: %v", err)
		}
	}

	b.StopTimer()

}