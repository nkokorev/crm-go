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

    user, err := account.CreateUser(User{Username: "TestUser_ExistUser", Phone: "88251001212", InvitedUserID:1, DefaultAccountID:1}, RoleAuthor)
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

}