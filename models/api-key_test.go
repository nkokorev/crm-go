package models

import (
	"github.com/nkokorev/crm-go/utils"
	"testing"
)


func TestGetApiKey(t *testing.T) {
	account, _ := Account{Name: "Test account for API Key"}.create()
	defer account.HardDelete()

	key, err := account.ApiKeyCreate(ApiKey{Name: "Api key for Postman"})
	if err != nil {
		t.Fatalf("Не удалось создать api-ключ для аккаунта: %v", err)
	}

	defer key.delete()

	sKey, err := account.ApiKeyGet(key.Id)
	if err != nil {
		t.Fatalf("Не удалось найти APiKey: %v", err)
	}
	if sKey == nil {
		t.Fatalf("Поиск по ключу вернул пустой указатель *")
	}
	if sKey.Token != key.Token {
		t.Fatal("При поиске токена по ключу найден какой-то другой ключ!")
	}
}

func TestApiKey_delete(t *testing.T) {
	account, _ := Account{Name: "Test account for API Key"}.create()
	defer account.HardDelete()

	key, err := account.ApiKeyCreate(ApiKey{Name: "Api key for Postman"})
	if err != nil {
		t.Fatalf("Не удалось создать api-ключ для аккаунта: %v", err)
	}

	// убеждаем, что сначала он его находит
	sKey, err := account.ApiKeyGet(key.Id)
	if err != nil || sKey == nil {
		t.Fatal("Ошибка с поиском ApiKey - он должен был найтись")
	}

	// удаляем ключ и затем проверим, что все работает
	err = key.delete()
	if err != nil {
		t.Fatalf("Не удалось удалить ApiKey")
	}

	_, err = account.ApiKeyGet(key.Id)
	if err == nil {
		t.Fatal("Найден apiKey, который был удален")
	}
}

func TestApiKey_update(t *testing.T) {

	account, _ := Account{Name: "Test account for API Key"}.create()
	defer account.HardDelete()

	key, _ := account.ApiKeyCreate(ApiKey{Name: "Api key for Test: " + utils.RandStringBytes(5)})
	defer key.delete()

	// Проверим, что новые данные сохраняются и не сохраняются лишние
	key.Name = utils.RandStringBytes(10) // должно сработать
	key.Enabled = !key.Enabled // должно сработать
	key.AccountId = key.AccountId + 1 // НЕ должно сработать
	key.Token = utils.RandStringBytes(10) // НЕ должно сработать

	/*if err := key.update(strings.(key)); err !=nil {
		t.Fatalf("Не удалось обновить ApiKey")
	}*/

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

