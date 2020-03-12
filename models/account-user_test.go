package models

import (
    "testing"
)

func Test_GetAccountUser(t *testing.T) {

    // создаем тестовый аккаунт, который потом удалим
    account, err := Account{Name:"TestAccount_GetUserRole"}.create()
    if err != nil {
        t.Fatalf("Неудалось создать тестовый аккаунт: %v", err)
    }
    defer account.HardDelete()

    // создаем тестового пользователя с ролью Автор
    user, err := account.CreateUser(User{Phone: "88251001212"}, RoleAuthor)
    if err !=nil || user == nil {
        t.Fatalf("Неудалось создать пользователя: %v", err)
    }
    defer user.hardDelete()

    // проверяем поиск пользователя aUser
    /*aUser, err := GetAccountUser(*account, *user)
    if err != nil {
        t.Fatalf("Неудалось найти пользователя aUser: %v", err)
    }
    if aUser == nil {
        t.Fatalf("Найденный пользователь == nil aUser: %v", aUser)
    }*/
    
}
