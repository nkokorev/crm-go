package models

import "testing"

func TestAccountUser_create(t *testing.T)  {
    account, err := Account{Name:"TestAccountUser_create"}.create()
    if err != nil {
        t.Fatalf("Неудалось создать тестовый аккаунт: %v\n", err)
    }
    defer func() {
        account.HardDelete()
    }()

    user, err := account.CreateUser(User{Username: "TestAccountUser_create", Phone: "88251001212", InvitedUserID:1, DefaultAccountID:1}, RoleAuthor)
    if err !=nil {
        t.Fatalf("Неудалось создать пользователя %v", err)
    }
    defer func() {
        user.hardDelete()
    }()

    role, err := GetRole(RoleManager)
    if err != nil || role == nil {
        t.Fatalf("Неудалось найти роль: %v\n", err)
    }

    testAUser_1 := AccountUser{
        AccountId:account.ID,
        UserId:user.ID,
        RoleId:role.ID,
    }

    aUser, err := testAUser_1.create();
    if err != nil || aUser == nil {
        t.Fatalf("Неудалось создать пользователя в аккаунте: %v\n", err)
    }
    defer func() {
        if err := aUser.delete(); err != nil {
            t.Fatalf("Не удалось удалить aUser: %v\n", err)
        }
    }()


    // а вот эти ниже не должны создаться
    testAUser_2 := AccountUser{
        AccountId: 373464,
        UserId:user.ID,
        RoleId:role.ID,
    }
    testAUser_3 := AccountUser{
        AccountId: account.ID,
        UserId:547854733,
        RoleId:role.ID,
    }
    testAUser_4 := AccountUser{
        AccountId: account.ID,
        UserId:user.ID,
        RoleId:33423,
    }

    aUser_2, err := testAUser_2.create();
    if err == nil{
        t.Fatalf("Удалось создать пользователя aUser, который не должен был пройти валидацию\n")
        aUser_2.delete();
    }
    aUser_3, err := testAUser_3.create();
    if err == nil {
        t.Fatalf("Удалось создать пользователя aUser, который не должен был пройти валидацию\n")
        aUser_3.delete();
    }
    aUser_4, err := testAUser_4.create();
    if err == nil{
        t.Fatalf("Удалось создать пользователя aUser, который не должен был пройти валидацию\n")
        aUser_4.delete();
    }



}

func TestAccountUser_update(t *testing.T)  {
    account, err := Account{Name:"TestAccountUser_update"}.create()
    if err != nil {
        t.Fatalf("Неудалось создать тестовый аккаунт: %v\n", err)
    }
    defer func() {
        account.HardDelete()
    }()

    user, err := account.CreateUser(User{Username: "TestAccountUser_update", Phone: "88251001248", InvitedUserID:1, DefaultAccountID:1}, RoleAdmin)
    if err !=nil {
        t.Fatalf("Неудалось создать пользователя %v", err)
    }
    defer func() {
        user.hardDelete()
    }()

    // Проверим, что пользователь создался
    aUser, err := account.GetAccountUser(*user)
    if err != nil || aUser == nil {
        t.Fatalf("Пользователь не был создан или добавлен в аккаунт: %v\n", err)
    }

    // Проверим, что это наш пользователь
    if aUser.User.Username != user.Username {
        t.Fatalf("Пользователь aUser получен с другими данными или вообще без них: %v\n", aUser.User)
    }
    if aUser.Role.Tag != RoleAdmin {
        t.Fatalf("Пользователь aUser получен с другими данными или вообще без них: %v\n", aUser.Role)
    }

    role, err := GetRole(RoleManager)
    if err != nil || role == nil {
        t.Fatalf("Не удалось получить роль: %v\n", err)
    }

    aUser.RoleId = role.ID
    if err := aUser.update(&aUser); err != nil {
        t.Fatalf("Не удалось обновить данные в aUser: %v\n", err)
    }

    // Проверим, что пользователь обновился
    aUser, err = account.GetAccountUser(*user)
    if err != nil || aUser == nil {
        t.Fatalf("Пользователь не был создан или добавлен в аккаунт: %v\n", err)
    }

    if aUser.RoleId != role.ID {
        t.Fatal("Данные aUser с ролью НЕ обновились!")
    }
}

func TestAccountUser_delete(t *testing.T) {
    // Создаем тестовый аккаунт
    account, err := Account{Name:"TestAccountUser_delete"}.create()
    if err != nil {
        t.Fatalf("Неудалось создать тестовый аккаунт: %v\n", err)
    }
    defer func() {
        account.HardDelete()
    }()

    user, err := account.CreateUser(User{Username: "TestAccountUser_delete", Phone: "88251009876", InvitedUserID:1, DefaultAccountID:1}, RoleAuthor)
    if err !=nil {
        t.Fatalf("Неудалось создать пользователя %v", err)
    }
    defer func() {
        user.hardDelete()
    }()

    aUser, err := account.GetAccountUser(*user)
    if err != nil || aUser == nil {
        t.Fatalf("Не удалось получить aUser или он *nil: %v\n", err)
    }

    // Удаляем через фукнцию, которую тестим
    if err := aUser.delete(); err != nil {
        t.Fatalf("Не удалось удалить aUser.delete(): %v\n", err)
    }

    aUser_2, err := account.GetAccountUser(*user)
    if err == nil || aUser_2 != nil {
        t.Fatalf("Удалось найти пользователя aUser после удаления: %v\n", aUser_2)
    }
}