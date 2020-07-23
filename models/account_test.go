package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"strings"
	"testing"
)

func TestAccount_ValidateInputs(t *testing.T) {

	account := Account{Name: ""}
	if err := account.ValidateInputs(); err == nil {
		t.Fatal("Validate account without name")
	}

	account.Name = utils.RandStringBytes(100)
	if err := account.ValidateInputs(); err == nil {
		t.Fatal("Validate account with very long name")
	}

	account.Name = utils.RandStringBytes(10)
	if err := account.ValidateInputs(); err != nil {
		t.Fatal("No validate account with short name")
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

	// 2. А вот этот аккаунт должен создаться
	outAccount, err := Account{Name: "Test account"}.create()
	if err != nil || outAccount == nil {
		fmt.Println(err)
		t.Fatal("Cant create account without name")
	}

	defer outAccount.HardDelete()

}

func TestGetAccount(t *testing.T) {
	// создаем тестовый аккаунт и потом его находим
	account, err := Account{Name: "Test Get account: " + utils.RandStringBytes(10)}.create()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый аккаунт: %v", err)
	}
	defer account.HardDelete()

	fAccount, err := GetAccount(account.Id)
	if err != nil {
		t.Fatal("Ошибка поиска аккаунта")
	}
	if fAccount == nil {
		t.Fatal("Не удалось найти вновь созданный аккаунт")
	}
	if fAccount.Name != account.Name {
		t.Fatalf("Поиск аккаунта дал не верный результат")
	}
}

func TestAccount_Exist(t *testing.T) {
	acc, err := GetMainAccount()
	if err != nil || acc == nil {
		t.Fatalf("Не удалось получить главный аккаунт: %v \n", err)
	}

	if !acc.Exist(acc.Id) {
		t.Fatal("Main аккаунт не существует, хотя на самом деле есть")
	}

	account, err := Account{Name: "TestAccount_Exist"}.create()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый аккаунт: %v", err)
	}
	defer account.HardDelete()

	if !account.Exist(account.Id) {
		t.Fatal("Тестовый аккаунт не существует, хотя на самом деле есть")
	}
}

func TestGetMainAccount(t *testing.T) {
	account, err := GetMainAccount()
	if err != nil || account.Id != 1 || account.Name != "RatusCRM" {
		t.Fatalf("Cant find main account: %v", err)
	}
}

func TestAccount_CreateApiKey(t *testing.T) {
	account, _ := Account{Name: "Test account for API Key"}.create()
	defer account.HardDelete()

	key, err := account.ApiKeyCreate(ApiKey{Name: "Api key for Test"})
	if err != nil {
		t.Fatalf("Не удалось создать api-ключ для аккаунта: %v", err)
	}

	defer key.delete()
}

func TestAccount_DeleteApiKey(t *testing.T) {
	account, _ := Account{Name: "Test account for API Key"}.create()
	defer account.HardDelete()

	key, err := account.ApiKeyCreate(ApiKey{Name: "Api key for Test"})
	if err != nil {
		t.Fatalf("Не удалось создать api-ключ для аккаунта: %v", err)
	}

	// убеждаем, что сначала он его находит
	sKey, err := account.ApiKeyGet(key.Id)
	if err != nil || sKey == nil {
		t.Fatal("Ошибка с поиском ApiKey - он должен был найтись")
	}

	// убеждаем, что нельзя удалить ключ из-под другого аккаунта
	account2, _ := Account{Name: "Test account for API Key 2"}.create()
	defer account2.HardDelete()

	err = account2.ApiKeyDelete(key.Id)
	if err == nil {
		t.Fatal("удалось удалить ApiKey из-под несвязанного аккаунта")
	}

	// а вот теперь должно удалиться
	err = account.ApiKeyDelete(key.Id)
	if err != nil {
		t.Fatalf("Не удалось удалить ApiKey: %v", err)
	}

	// убеждаемся, что после удаления нашего ключика нет
	_, err = account.ApiKeyGet(key.Id)
	if err == nil {
		t.Fatal("Найден apiKey, который был удален")

	}
}

func TestAccount_GetApiKey(t *testing.T) {

	account, _ := Account{Name: "Test account for API Key"}.create()
	defer account.HardDelete()

	account2, _ := Account{Name: "Test account for API Key 2"}.create()
	defer account2.HardDelete()

	key, _ := account.ApiKeyCreate(ApiKey{Name: "Api key for Test"})
	defer account.ApiKeyDelete(key.Id)

	// убеждаем, что нельзя получить ключ из-под другого аккаунта
	_, err := account2.ApiKeyGet(key.Id)
	if err == nil {
		t.Fatal("удалось получить ApiKey из-под несвязанного аккаунта")
	}
}

func TestAccount_UpdateApiKey(t *testing.T) {
	account, _ := Account{Name: "Test account for API Key"}.create()
	defer account.HardDelete()

	key, _ := account.ApiKeyCreate(ApiKey{Name: "Api key for Test: " + utils.RandStringBytes(5)})
	defer account.ApiKeyDelete(key.Id)

	// Проверим, что новые данные сохраняются и не сохраняются лишние
	// token := key.Token
	key.Name = utils.RandStringBytes(10) // должно сработать
	key.Enabled = !key.Enabled // должно сработать
	key.AccountId = key.AccountId + 1 // НЕ должно сработать
	key.Token = utils.RandStringBytes(10) // НЕ должно сработать

	if err := key.update(*key); err !=nil {
		t.Fatalf("Не удалось обновить ApiKey")
	}

	sKey, err := account.ApiKeyGet(key.Id)
	if err != nil {
		t.Fatal("Не удалось найти ApiKey после update")
	}

	if sKey.Token == key.Token {
		t.Fatal("Удалось обновлением изменить token у ApiKey")
	}
	if sKey.Enabled != key.Enabled {
		t.Fatal("Удалось обновлением изменить Enabled у ApiKey")
	}
	if sKey.AccountId != account.Id {
		t.Fatal("Удалось обновлением изменить AccountId у ApiKey")
	}
}

func TestAccount_CreateUser(t *testing.T) {

	// аккаунт с регистрацией по имени пользователя
	account, err := Account{Name: "Test account CreateUser Username"}.create()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый аккаунт: %v", err)
	}
	defer account.HardDelete()

	// todo дописать список тестов
	testList := []struct {
		account     *Account
		user        User
		accessRole  accessRole
		expected    bool
		description string
		}{
			{account, User{Username: "", Email:"", Phone:""}, RoleAuthor, false, "Хотя бы один из определяющих полей должен быть"},
			{account, User{Username: "TestUser 1", Email:"adnsls!@.ru"}, RoleAdmin, false, "Не корректный email"},
			{account, User{Username: "TestUser 1", Phone:"5456a45355"}, RoleClient, false, "Не корректный телефон"},
		}


	for i, v := range testList {
		user, err := v.account.CreateUser(v.user, v.accessRole)

		if v.expected == false && err == nil {
			t.Fatalf("Создан пользователь, которого быть не должно : [%v] user: %v", i, user)
		}

		if v.expected == true && err != nil {
			t.Fatalf("Не удалось создать пользователя, который должен быть создан: [%v] user: %v", i, user)
		}
		
		// Проверяем роль и удаляем созданного пользователя
		if err == nil && user != nil {

			accessRole, err := v.account.GetUserAccessRole(*user)
			if err != nil || accessRole == nil {
				t.Fatalf("Не удалось получить роль пользователя [%v] : %v", user.Name, err)
			}
			if *accessRole != v.accessRole {
				t.Fatalf("Роль пользователя не соответствует ожидаемой: [%v] user: %v", v.accessRole, user)
			}

			user.hardDelete()
		}

	}
}

func TestAccount_GetUserById(t *testing.T) {
	account, err := Account{Name: "TestAccount_GetUserById"}.create()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый аккаунт: %v", err)
	}
	defer account.HardDelete()

	user, err := account.CreateUser(User{Username: "TestUser"})
	if err!=nil {
		t.Fatalf("Не удалось создать пользователя %v", err)
	}
	defer user.hardDelete()

	userF, err := account.GetUserById(user.Id)
	if err != nil {
		t.Fatalf("Не удалось найти пользователя, %v", err)
	}

	if userF.Id != user.Id || userF.Id == 0 {
		t.Fatalf("Ошибка: пользователь найден не правильно!")
	}

}

func TestAccount_GetUserByUsername(t *testing.T) {
	account, err := Account{Name: "TestAccount_GetUserById"}.create()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый аккаунт: %v", err)
	}
	defer account.HardDelete()

	user, err := account.CreateUser(User{Username: utils.RandStringBytes(10)})
	if err!=nil {
		t.Fatalf("Не удалось создать пользователя %v", err)
	}
	defer user.hardDelete()

	fUser, err := account.GetUserByUsername(user.Username)
	if err != nil {
		t.Fatalf("Не удалось найти пользователя, %v", err)
	}

	if fUser.Id != user.Id || fUser.Id == 0 || fUser.Username != user.Username {
		t.Fatalf("Ошибка: пользователь найден не правильно!")
	}

}

func TestAccount_GetUserByEmail(t *testing.T) {
	account, err := Account{Name: "TestAccount_GetUserById"}.create()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый аккаунт: %v", err)
	}
	defer account.HardDelete()

	user, err := account.CreateUser(User{Email: strings.ToLower(utils.RandStringBytes(5)) + "@rus-marketing.com"})
	if err!=nil {
		t.Fatalf("Не удалось создать пользователя %v", err)
	}
	defer user.hardDelete()

	fUser, err := account.GetUserByEmail(user.Email)
	if err != nil {
		t.Fatalf("Не удалось найти пользователя, %v", err)
	}

	if fUser.Id != user.Id || fUser.Id == 0 || fUser.Email != user.Email {
		t.Fatalf("Ошибка: пользователь найден не правильно!")
	}

}

func TestAccount_GetUserByPhone(t *testing.T) {
	account, err := Account{Name: "TestAccount_GetUserByPhone"}.create()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый аккаунт: %v", err)
	}
	defer account.HardDelete()

	user, err := account.CreateUser(User{Phone: "88251001212"})
	if err!=nil {
		t.Fatalf("Не удалось создать пользователя %v", err)
	}
	defer user.hardDelete()

	fUser, err := account.GetUserByPhone(user.Phone, "")
	if err != nil {
		t.Fatalf("Не удалось найти пользователя, %v", err)
	}

	if fUser.Id != user.Id || fUser.Id == 0 || fUser.Phone != user.Phone {
		t.Fatalf("Ошибка: пользователь найден не правильно!")
	}

}

func TestAccount_GetAccountUser(t *testing.T) {
	account1, err := Account{Name: "TestAccount_GetAccountUser_1"}.create()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый аккаунт: %v", err)
	}
	defer func() {
		account1.HardDelete()
	}()

	account2, err := Account{Name: "TestAccount_GetAccountUser_2"}.create()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый аккаунт: %v", err)
	}
	defer func() {
		account2.HardDelete()
	}()

	// создаем тестового пользователя с ролью Автор
	user, err := account1.CreateUser(User{Username: "GetAccountUser", Phone: "89251251001534", InvitedUserId:1}, RoleClient)
	if err!=nil {
		t.Fatalf("Не удалось создать пользователя %v", err)
	}
	defer func() {
		user.hardDelete()
	}()

	// тестируем саму функцию GetAccountUser
	aUser, err := account1.GetAccountUser(*user)
	if err  != nil && err != gorm.ErrRecordNotFound {
		t.Fatalf("Ошибка при поиске aUser: %v", err)
	}
	if err == gorm.ErrRecordNotFound {
		t.Fatal("Пользователь не найден, хотя должен был бы...")
	}
	if aUser == nil {
		t.Fatalf("Вместо пользователя найден *nil: %v", aUser)
	}
}

func TestAccount_ExistUser(t *testing.T) {

	account1, err := Account{Name: "TestAccount_AppendUser_1"}.create()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый аккаунт: %v", err)
	}
	defer func() {
		account1.HardDelete()
	}()

	account2, err := Account{Name: "TestAccount_AppendUser_2"}.create()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый аккаунт: %v", err)
	}
	defer func() {
		account2.HardDelete()
	}()

	// создаем тестового пользователя с ролью Автор
	user, err := account1.CreateUser(User{Username: "TestUser_ExistUser", Phone: "88251001212", InvitedUserId:1}, RoleAuthor)
	if err!=nil {
		t.Fatalf("Не удалось создать пользователя %v", err)
	}
	defer func() {
		user.hardDelete()
	}()

	// проверяем есть ли вообще указанный пользователь
	if !(Account{}).ExistUser(*user) {
		t.Fatal("Не найден пользователь, который должен быть")
	}
}

func TestAccount_ExistAccountUser(t *testing.T) {
	account1, err := Account{Name: "TestAccount_AppendUser_1"}.create()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый аккаунт: %v", err)
	}
	defer func() {
		account1.HardDelete()
	}()

	account2, err := Account{Name: "TestAccount_AppendUser_2"}.create()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый аккаунт: %v", err)
	}
	defer func() {
		account2.HardDelete()
	}()

	// создаем тестового пользователя с ролью Автор
	user, err := account1.CreateUser(User{Username: "TestUser_ExistUser", Phone: "88251001212", InvitedUserId:1}, RoleAuthor)
	if err!=nil {
		t.Fatalf("Не удалось создать пользователя %v", err)
	}
	defer func() {
		user.hardDelete()
	}()

	if !account1.ExistAccountUser(*user) {
		t.Fatal("Не найден пользователь в аккаунте 1, который должен быть")
	}

	if account2.ExistAccountUser(*user) {
		t.Fatal("Найден пользователь в аккаунте 2, которого не должно было быть")
	}
}

func TestAccount_existUserByUsername(t *testing.T) {
	account, err := Account{Name: "TestAccount_existUserUsername"}.create()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый аккаунт: %v", err)
	}
	defer account.HardDelete()

	account2, err := Account{Name: "TestAccount_existUserUsername_2"}.create()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый аккаунт: %v", err)
	}
	defer func() {
		account2.HardDelete()
	}()

	// создаем тестового пользователя с ролью Автор
	user, err := account.CreateUser(User{Username: "TestUser1234",Phone: "+79251958873"}, RoleClient)
	if err!=nil {
		t.Fatalf("Не удалось создать пользователя: %v", err)
	}
	defer user.hardDelete()

	// Проверим функцию
	if !account.existUserByUsername(user.Username) {
		t.Fatal("Не удалось найти пользователя, который должен быть")
	}
	if account2.existUserByUsername(user.Username) {
		t.Fatal("Удалось найти пользователя, которого не должно быть")
	}
}

func TestAccount_existUserByEmail(t *testing.T) {
	account, err := Account{Name: "TestAccount_existUserByEmail"}.create()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый аккаунт: %v", err)
	}
	defer account.HardDelete()

	account2, err := Account{Name: "TestAccount_existUserByEmail_2"}.create()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый аккаунт: %v", err)
	}
	defer func() {
		account2.HardDelete()
	}()

	// создаем тестового пользователя с ролью Автор
	user, err := account.CreateUser(User{Email: "testmail@ratus-dev.ru"}, RoleClient)
	if err!=nil {
		t.Fatalf("Не удалось создать пользователя: %v", err)
	}
	defer user.hardDelete()

	// Проверим функцию
	if !account.existUserByEmail(user.Email) {
		t.Fatal("Не удалось найти пользователя, который должен быть")
	}
	if account2.existUserByEmail(user.Email) {
		t.Fatal("Удалось найти пользователя, которого не должно быть")
	}
}

func TestAccount_existUserByPhone(t *testing.T) {
	account, err := Account{Name: "TestAccount_existUserUsername"}.create()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый аккаунт: %v", err)
	}
	defer account.HardDelete()

	account2, err := Account{Name: "TestAccount_existUserUsername_2"}.create()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый аккаунт: %v", err)
	}
	defer func() {
		account2.HardDelete()
	}()

	// создаем тестового пользователя с ролью Автор
	user, err := account.CreateUser(User{Phone: "+79997775554411"}, RoleClient)
	if err!=nil {
		t.Fatalf("Не удалось создать пользователя: %v", err)
	}
	defer user.hardDelete()

	// Проверим функцию
	if !account.existUserByPhone(user.Phone) {
		t.Fatal("Не удалось найти пользователя, который должен быть")
	}
	if account2.existUserByPhone(user.Phone) {
		t.Fatal("Удалось найти пользователя, которого не должно быть")
	}
}

func TestAccount_AppendUser(t *testing.T) {
	
	account, err := Account{Name: "TestAccount_AppendUser_1"}.create()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый аккаунт: %v", err)
	}
	defer func() {
		account.HardDelete()
	}()

	account2, err := Account{Name: "TestAccount_AppendUser_2"}.create()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый аккаунт: %v", err)
	}
	defer func() {
		account2.HardDelete()
	}()

	// создаем тестового пользователя с ролью Автор
	user, err := account.CreateUser(User{Phone: "88251001018"}, RoleAuthor)
	if err!=nil || user == nil {
		t.Fatalf("Не удалось создать пользователя %v", err)
	}
	defer func() {
		user.hardDelete()
	}()

	if !account.ExistAccountUser(*user) {
		t.Fatal("Пользователь не был добавлен функцией account.CreateUser() в аккаунт_1.")
	}

	if account2.ExistAccountUser(*user) {
		t.Fatal("Пользователь был добавлен функцией account.CreateUser() в аккаунт_2.")
	}

	aUser, err := account2.AppendUser(*user, RoleClient);
	if err != nil || aUser == nil{
		t.Fatalf("Не удалось добавить пользователя в аккаунт с ролью Клиент: %v", err)
	}

	aRole, err := account2.GetUserAccessRole(*user)
	if err != nil {
		t.Fatalf("Не удалось получить AccessRole: %v", err)
	}
	if aRole == nil {
		t.Fatalf("Роль пользователя нулевая (не найдена): %v", aRole)
	}
}

func TestAccount_GetUserRole(t *testing.T) {

	// создаем тестовый аккаунт
	account, err := Account{Name: "TestAccount_GetUserRole"}.create()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый аккаунт: %v", err)
	}
	defer account.HardDelete()

	// создаем тестового пользователя с ролью Автор
	user, err := account.CreateUser(User{Phone: "88251001212"}, RoleClient)
	if err!=nil {
		t.Fatalf("Не удалось создать пользователя %v", err)
	}
	defer user.hardDelete()

	// проверяем роль нашего пользователя
	role, err := account.GetUserRole(*user)
	if err != nil || role == nil {
		t.Fatalf("Не удалось получить роль пользователя: %v", err)
	}

	// сверяем полученные роли
	if role.Tag != RoleClient {
		t.Fatalf("Роль созданного пользователя не является Client: %v", role)
	}
	
}

func TestAccount_GetUserAccessRole(t *testing.T) {
	// создаем тестовый аккаунт
	account, err := Account{Name: "TestAccount_GetUserAccessRole"}.create()
	if err != nil {
		t.Fatalf("Не удалось создать тестовый аккаунт: %v", err)
	}
	defer account.HardDelete()

	// создаем тестового пользователя с ролью Автор
	user, err := account.CreateUser(User{Phone: "88251001212"}, RoleClient)
	if err!=nil {
		t.Fatalf("Не удалось создать пользователя: %v", err)
	}
	defer user.hardDelete()

	// Проверяем роль нашего пользователя
	accessRole, err := account.GetUserAccessRole(*user)
	if err != nil || accessRole == nil {
		t.Fatalf("Не удалось получить роль пользователя: %v", err)
	}

	if *accessRole != RoleClient {
		t.Fatalf("Роль созданного пользователя не является Автор: %v.", *accessRole)
	}
}



// #### Benchmark Go! ####

func BenchmarkGetAccountByHash(b *testing.B) {
	// создаем много аккаунтов
	return
	/*runAccounts := 50000
	for i:=0;i < runAccounts;i++ {
		_, err := Account{Name:"TestAccount"}.create()
		if err != nil {
			b.Fatalf("Не удалось создать аккаунт, %v", err)
		}
		//defer account.HardDelete()
	}
	return*/
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		acc, err := GetAccountByHash("arfjdlafkl")
		if err != nil || acc == nil {
			b.Fatalf("Не удалось найти аккаунт: %v", err)
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
			b.Fatalf("Не удалось найти аккаунт: %v", err)
		}
	}

	b.StopTimer()

}