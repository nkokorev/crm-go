package models

import "testing"

// ### Testing CRUD functions

func TestUser_create(t *testing.T) {
	// see: TestAccount_UserCreate in account_test.go
}

func TestUser_getUserById(t *testing.T)  {
	account, err := Account{Name:"TestAccount_getUserById"}.create()
	if err != nil {
		t.Fatalf("Неудалось создать тестовый аккаунт: %v", err)
	}
	defer account.HardDelete()

	user := User{
		SignedAccountID:account.ID,
		Username:"TestUser_getUserById",
	}
	u, err := user.create()
	if err !=nil {
		t.Fatalf("Не удалось создать пользователя, %v", err)
	}
	defer u.hardDelete()
}

func TestUser_hardDelete(t *testing.T)  {
	account, err := Account{Name:"TestAccount_hardDelete"}.create()
	if err != nil {
		t.Fatalf("Неудалось создать тестовый аккаунт: %v", err)
	}
	defer account.HardDelete()

	testUser := &User{
		SignedAccountID:account.ID,
		Username:"TestUser_getUserById",
	}
	user, err := testUser.create()
	if err !=nil {
		t.Fatalf("Не удалось создать пользователя, %v", err)
	}
	if err := user.hardDelete(); err != nil {
		t.Fatalf("Не удалось удалить пользователя, %v", err)
	}
	// проверяем, что пользователя нет
	_, err = getUserById(user.ID)
	if err == nil {
		t.Fatalf("Пользователь на самом деле не удалился")
	}

}

func TestUser_softDelete(t *testing.T)  {
	account, err := Account{Name:"TestUser_softDelete"}.create()
	if err != nil {
		t.Fatalf("Неудалось создать тестовый аккаунт: %v", err)
	}
	defer account.HardDelete()

	testUser := &User{
		SignedAccountID:account.ID,
		Username:"TestUser_getUserById",
	}
	user, err := testUser.create()
	if err !=nil {
		t.Fatalf("Не удалось создать пользователя, %v", err)
	}
	defer user.hardDelete()

	if err := user.softDelete(); err != nil {
		t.Fatalf("Не удалось удалить пользователя, %v", err)
	}

	// проверяем, что пользователя нет
	_, err = getUserById(user.ID)
	if err == nil {
		t.Fatalf("Пользователь на самом деле не удалился")
	}

	// а вот тут пользователь должен был найтись
	fUser, err := getUnscopedUserById(user.ID)
	if err != nil {
		t.Fatalf("Пользователь должен был найтись")
	}

	if fUser.ID != user.ID {
		t.Fatalf("ID пользователей не совпадают")
	}

}


